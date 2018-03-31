package dht

import (
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
)

type Pair struct {
	Key        string            `json:"k"`
	Value      string            `json:"v"`
	Labels     map[string]string `json:"l"`
	LastPut    time.Time         `json:"lp"`
	Persistent bool              `json:"i"`
}

func (p Pair) GetKey() string {
	return p.Key
}

func (p Pair) GetValue() string {
	return p.Value
}

func (p Pair) GetLabels() map[string]string {
	return p.Labels
}

func (p Pair) GetLabel(key string) string {
	return p.Labels[key]
}

type Store struct {
	pairs map[string][]Pair
	lock  *sync.RWMutex
}

func newStore() (*Store, error) {
	s := &Store{
		pairs: map[string][]Pair{},
		lock:  &sync.RWMutex{},
	}
	return s, nil
}

func (s *Store) Put(key, value string, labels map[string]string, persistent bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// make sure our partition exists
	if _, ok := s.pairs[key]; !ok {
		s.pairs[key] = []Pair{}
	}

	// check if the pair already exists
	for _, pair := range s.pairs[key] {
		if pair.Value == value && reflect.DeepEqual(pair.Labels, labels) {
			// and if it does update it
			pair.LastPut = time.Now()
			return nil
		}
	}

	// else add it
	pair := Pair{
		Key:        key,
		Value:      value,
		Labels:     labels,
		LastPut:    time.Now(),
		Persistent: persistent,
	}
	s.pairs[key] = append(s.pairs[key], pair)

	return nil
}

func (s *Store) Get(key string) ([]Pair, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	rpairs := []Pair{}

	// check if our partition exists
	pairs, ok := s.pairs[key]
	if !ok {
		logrus.WithField("key", key).Debugf("store.Get new partition")
		return rpairs, nil
	}

	// get call values
	for _, pair := range pairs {
		rpairs = append(rpairs, pair)
	}

	logrus.WithField("pairs", rpairs).WithField("key", key).Debugf("store.Get")

	return rpairs, nil
}

// TODO filter only supports a single label for now
func (s *Store) Filter(key string, labels map[string]string) ([]Pair, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	rpairs := []Pair{}

	// check if our partition exists
	pairs, ok := s.pairs[key]
	if !ok {
		logrus.WithField("key", key).Debugf("store.Get new partition")
		return rpairs, nil
	}

	// get call values
	for _, pair := range pairs {
		if len(labels) == 0 {
			rpairs = append(rpairs, pair)
			continue
		}
		for key, value := range labels {
			if pair.Labels[key] == value {
				rpairs = append(rpairs, pair)
				break
			}
		}
	}

	logrus.WithField("pairs", rpairs).WithField("key", key).Debugf("store.Get")

	return rpairs, nil
}

// FindPeersNearestTo returns an array of n peers closest to the given key by xor distance
func (s *Store) FindPeersNearestTo(tk string, n int) ([]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// place to hold the results
	rks := []string{}

	htk := hash(tk)

	// slice to hold the distances
	dists := []distEntry{}
	for ik, pairs := range s.pairs {
		for _, pair := range pairs {
			if pair.GetLabel("protocol") != "messaging" {
				continue
			}
			// calculate distance
			de := distEntry{
				key:  ik,
				dist: xor([]byte(htk), []byte(hash(ik))),
			}
			exists := false
			for _, ee := range dists {
				if ee.key == pair.Key {
					exists = true
					break
				}
			}
			if !exists {
				dists = append(dists, de)
			}
		}
	}

	// sort the distances
	sort.Slice(dists, func(i, j int) bool {
		return lessIntArr(dists[i].dist, dists[j].dist)
	})

	if n > len(dists) {
		n = len(dists)
	}

	// append n the first n number of keys
	for _, de := range dists {
		rks = append(rks, de.key)
		n--
		if n == 0 {
			break
		}
	}

	return rks, nil
}

func (s *Store) Wipe(key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.pairs[key]; !ok {
		return nil
	}

	npr := []Pair{}
	for _, pr := range s.pairs[key] {
		if pr.Persistent {
			npr = append(npr, pr)
		}
	}

	if len(npr) == 0 {
		delete(s.pairs, key)
		return nil
	}

	s.pairs[key] = npr
	return nil
}

func (s *Store) GetAll() (map[string][]Pair, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return deepcopy.Copy(s.pairs).(map[string][]Pair), nil
}
