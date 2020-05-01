package hyperspace

import (
	"sync"

	"nimona.io/pkg/object"
)

type (
	pinnedObjects struct {
		m sync.Map
	}
)

// Put -
func (m *pinnedObjects) Put(k object.Hash) {
	m.m.Store(k, true)
}

// Delete -
func (m *pinnedObjects) Delete(k object.Hash) {
	m.m.Delete(k)
}

// List -
func (m *pinnedObjects) List() []object.Hash {
	hs := []object.Hash{}
	m.m.Range(func(k, v interface{}) bool {
		hs = append(hs, k.(object.Hash))
		return true
	})
	return hs
}
