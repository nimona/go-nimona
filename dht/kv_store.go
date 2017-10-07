package dht

import (
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Pair struct {
	Key        string    `json:"k"`
	Value      string    `json:"v"`
	LastPut    time.Time `json:"lp"`
	Persistent bool      `json:"i"`
}

type Store struct {
	pairs map[string][]*Pair
	lock  *sync.RWMutex
}

func newStore() (*Store, error) {
	s := &Store{
		pairs: map[string][]*Pair{},
		lock:  &sync.RWMutex{},
	}
	return s, nil
}

func (s *Store) Put(key, value string, persistent bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// make sure our partition exists
	if _, ok := s.pairs[key]; !ok {
		s.pairs[key] = []*Pair{}
	}

	// check if the pair already exists
	for _, pair := range s.pairs[key] {
		if pair.Value == value {
			// and if it does update it
			pair.LastPut = time.Now()
			return nil
		}
	}

	// else add it
	pair := &Pair{
		Key:        key,
		Value:      value,
		LastPut:    time.Now(),
		Persistent: persistent,
	}
	s.pairs[key] = append(s.pairs[key], pair)

	return nil
}

func (s *Store) Get(key string) ([]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// check if our partition exists
	pairs, ok := s.pairs[key]
	if !ok {
		logrus.WithField("key", key).Debugf("store.Get new partition")
		return []string{}, nil
	}

	// get call values
	vs := []string{}
	for _, pair := range pairs {
		vs = append(vs, pair.Value)
	}

	logrus.WithField("vs", vs).WithField("key", key).Debugf("store.Get")

	return vs, nil
}

// FindKeysNearestTo returns an array of n keys closest to the given key by xor distance
func (s *Store) FindKeysNearestTo(prefix, tk string, n int) ([]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// place to hold the results
	rks := []string{}

	for key := range s.pairs {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		rks = append(rks, key)
		if len(rks) == n {
			break
		}
	}

	return rks, nil

	// htk := hash(tk)

	// // slice to hold the distances
	// dists := []distEntry{}
	// for ik := range s.pairs {
	// 	// only keep correct prefixe
	// 	if !strings.HasPrefix(ik, prefix) {
	// 		continue
	// 	}
	// 	// calculate distance
	// 	de := distEntry{
	// 		key:  ik,
	// 		dist: xor([]byte(htk), []byte(hash(ik))),
	// 	}
	// 	dists = append(dists, de)
	// }

	// // sort the distances
	// sort.Slice(dists, func(i, j int) bool {
	// 	return lessIntArr(dists[i].dist, dists[j].dist)
	// })

	// if n > len(dists) {
	// 	n = len(dists)
	// }

	// // append n the first n number of keys
	// for _, de := range dists {
	// 	rks = append(rks, de.key)
	// 	n--
	// 	if n == 0 {
	// 		break
	// 	}
	// }

	// return rks, nil
}

func (s *Store) Wipe(key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.pairs[key] = []*Pair{}
	return nil
}

func (s *Store) GetAll() (map[string][]*Pair, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pairs, nil
}
