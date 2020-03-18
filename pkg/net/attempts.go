package net

import (
	"sync"
)

type (
	attemptsMap struct {
		m sync.Map
	}
)

// newAttemptsMap constructs a new SyncMap
func newAttemptsMap() *attemptsMap {
	return &attemptsMap{}
}

// GetOrPut -
func (m *attemptsMap) GetOrPut(k string, v int) (int, bool) {
	nv, ok := m.m.LoadOrStore(k, v)
	return nv.(int), ok
}

// Put -
func (m *attemptsMap) Put(k string, v int) {
	m.m.Store(k, v)
}

// Get -
func (m *attemptsMap) Get(k string) (int, bool) {
	i, ok := m.m.Load(k)
	if !ok {
		return 0, false
	}

	v, ok := i.(int)
	if !ok {
		return 0, false
	}

	return v, true
}
