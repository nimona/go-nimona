package kv

import (
	"container/list"
	"strings"
	"sync"
)

type (
	lru struct {
		c int
		l *list.List
		m sync.Map
	}
	lruElement struct {
		element *list.Element
		value   []byte
	}
)

// NewLRU constrcuts an in-memory LRU key-value store
func NewLRU(capacity int) Store {
	return &lru{
		c: capacity,
		l: &list.List{},
	}
}

// Put a key-value pair, if it didn't already exist remove the last element
// from the linked list
func (m *lru) Put(k string, v []byte) error {
	if v, ok := m.m.Load(k); ok {
		e := v.(*lruElement)
		m.l.MoveToFront(e.element)
		return nil
	}

	e := &lruElement{
		element: m.l.PushFront(k),
		value:   v,
	}

	m.m.Store(k, e)
	if m.l.Len() > m.c {
		le := m.l.Back()
		lk := le.Value.(string)
		m.m.Delete(lk)
	}
	return nil
}

// Get the value of a key
func (m *lru) Get(k string) ([]byte, error) {
	v, ok := m.m.Load(k)
	if !ok {
		return nil, ErrNotFound
	}

	e, ok := v.(*lruElement)
	if !ok {
		return nil, ErrNotFound
	}

	m.l.MoveToFront(e.element)
	return e.value, nil
}

// Remove the value of a key
func (m *lru) Remove(k string) error {
	v, ok := m.m.Load(k)
	if !ok {
		return nil
	}

	e := v.(*lruElement)
	m.l.Remove(e.element)
	m.m.Delete(k)
	return nil
}

// List all keys
func (m *lru) List() ([]string, error) {
	ks := []string{}
	m.m.Range(func(k, v interface{}) bool {
		ks = append(ks, k.(string))
		return true
	})
	return ks, nil
}

// Scan for a key prefix and return all matching keys
func (m *lru) Scan(prefix string) ([]string, error) {
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
