package dht

import (
	"sync"

	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
)

type Store struct {
	values    map[string]string
	providers map[string][]string
	lock      *sync.RWMutex
}

func newStore() (*Store, error) {
	s := &Store{
		values:    map[string]string{},
		providers: map[string][]string{},
		lock:      &sync.RWMutex{},
	}
	return s, nil
}

func (s *Store) PutProvider(key string, providers ...string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// make sure our partition exists
	if _, ok := s.providers[key]; !ok {
		s.providers[key] = []string{}
	}

	for _, provider := range providers {
		// check if the pair already exists
		for _, existingProvider := range s.providers[key] {
			if existingProvider == provider {
				continue
			}
		}
		// else add it
		s.providers[key] = append(s.providers[key], provider)
	}

	return nil
}

func (s *Store) GetProviders(key string) ([]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// check if our partition exists
	providers, ok := s.providers[key]
	if !ok {
		logrus.WithField("key", key).Debugf("store.Get new partition")
		return []string{}, nil
	}

	logrus.WithField("providers", providers).WithField("key", key).Debugf("store.Get")
	return providers, nil
}

func (s *Store) PutValue(key string, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.values[key] = value
	return nil
}

func (s *Store) GetValue(key string) (string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.values[key], nil
}

func (s *Store) Wipe(key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.providers = map[string][]string{}

	return nil
}

func (s *Store) GetAllProviders() (map[string][]string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return deepcopy.Copy(s.providers).(map[string][]string), nil
}

func (s *Store) GetAllValues() (map[string]string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return deepcopy.Copy(s.values).(map[string]string), nil
}
