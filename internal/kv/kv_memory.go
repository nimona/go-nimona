package kv

import (
	"strings"

	"nimona.io/internal/xsync"
)

func NewMemoryStore[K, V any]() Store[K, V] {
	return &Memory[K, V]{
		items: &xsync.Map[string, V]{},
	}
}

type Memory[K, V any] struct {
	items *xsync.Map[string, V]
}

func (s *Memory[K, V]) Set(key K, value *V) error {
	s.items.Store(keyToString(key), *value)
	return nil
}

func (s *Memory[K, V]) Get(key K) (*V, error) {
	value, ok := s.items.Load(keyToString(key))
	if !ok {
		return nil, nil
	}
	return &value, nil
}

func (s *Memory[K, V]) GetPrefix(key K) ([]*V, error) {
	prefix := keyToString(key)
	values := []*V{}
	s.items.Range(func(k string, v V) bool {
		if strings.HasPrefix(k, prefix) {
			values = append(values, &v)
		}
		return true
	})
	return values, nil
}
