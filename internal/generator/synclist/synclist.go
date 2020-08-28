package syncmap

import (
	"sync"

	"github.com/geoah/genny/generic"
)

type (
	KeyType generic.Type // nolint
	// KeyTypeSyncList -
	KeyTypeSyncList struct {
		m sync.Map
	}
)

// Put -
func (m *KeyTypeSyncList) Put(k KeyType) {
	m.m.Store(k, true)
}

// Exists -
func (m *KeyTypeSyncList) Exists(k KeyType) bool {
	_, ok := m.m.Load(k)
	return ok
}

// Delete -
func (m *KeyTypeSyncList) Delete(k KeyType) {
	m.m.Delete(k)
}

// Range -
func (m *KeyTypeSyncList) Range(i func(k KeyType) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return i(k.(KeyType))
	})
}

// List -
func (m *KeyTypeSyncList) List() []KeyType {
	r := []KeyType{}
	m.m.Range(func(k, v interface{}) bool {
		r = append(r, k.(KeyType))
		return true
	})
	return r
}
