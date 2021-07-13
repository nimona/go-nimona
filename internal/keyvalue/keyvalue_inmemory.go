package keyvalue

import "sync"

// NewInMemoryStore returns a new in-memory, thread-safe key-value store.
func NewInMemoryStore() Store {
	return &inMemoryStore{
		mutex:  sync.RWMutex{},
		values: map[string][]byte{},
	}
}

// inMemoryStore is an in-memory, thread-safe key-value store.
type inMemoryStore struct {
	mutex  sync.RWMutex
	values map[string][]byte
}

// Get gets the value for the given key.
// If the key does not exist, then Get returns ErrNotFound.
func (s *inMemoryStore) Get(key string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, ok := s.values[key]
	if !ok {
		return nil, ErrNotFound
	}
	return value, nil
}

// Set sets the value for the given key.
func (s *inMemoryStore) Set(key string, value []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.values[key] = value
	return nil
}

// Delete deletes the value for the given key.
func (s *inMemoryStore) Delete(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.values, key)
	return nil
}

// Iter calls f for each key-value pair in the store.
func (s *inMemoryStore) Iter(f func(key string, value []byte)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for k, v := range s.values {
		f(k, v)
	}
}

// Close closes the store.
func (s *inMemoryStore) Close() error {
	return nil
}
