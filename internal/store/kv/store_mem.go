package kv

import (
	"strings"
	"sync"
)

type mem struct {
	m sync.Map
}

// NewMemory constrcuts an in-memory key-valye store
func NewMemory() Store {
	return &mem{}
}

// Put a key-value pair
func (m *mem) Put(k string, v []byte) error {
	if _, ok := m.m.Load(k); ok {
		return ErrExists
	}

	m.m.Store(k, v)
	return nil
}

// Get the value of a key
func (m *mem) Get(k string) ([]byte, error) {
	v, ok := m.m.Load(k)
	if !ok {
		return nil, ErrNotFound
	}

	b, ok := v.([]byte)
	if !ok {
		return nil, ErrNotFound
	}

	return b, nil
}

// Remove the value of a key
func (m *mem) Remove(k string) error {
	m.m.Delete(k)
	return nil
}

// List all keys
func (m *mem) List() ([]string, error) {
	ks := []string{}
	m.m.Range(func(k, v interface{}) bool {
		ks = append(ks, k.(string))
		return true
	})
	return ks, nil
}

// Scan for a key prefix and return all matching keys
func (m *mem) Scan(prefix string) ([]string, error) {
	ks := []string{}
	m.m.Range(func(k, v interface{}) bool {
		key := k.(string)
		if strings.HasPrefix(key, prefix) {
			ks = append(ks, key)
		}
		return true
	})
	return ks, nil
}
