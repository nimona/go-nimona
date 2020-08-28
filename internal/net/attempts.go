package net

import (
	"sync"
)

type attemptsMap struct {
	sync.Map
}

// GetOrPut -
func (m *attemptsMap) GetOrPut(k string, v int) (int, bool) {
	nv, ok := m.LoadOrStore(k, v)
	return nv.(int), ok
}

// Put -
func (m *attemptsMap) Put(k string, v int) {
	m.Store(k, v)
}

// Get -
func (m *attemptsMap) Get(k string) (int, bool) {
	i, ok := m.Load(k)
	if !ok {
		return 0, false
	}

	v, ok := i.(int)
	if !ok {
		return 0, false
	}

	return v, true
}
