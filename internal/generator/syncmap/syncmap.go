package syncmap

import (
	"sync"

	"github.com/cheekybits/genny/generic"
)

type (
	KeyType   generic.Type // nolint
	ValueType generic.Type // nolint
	// KeyTypeValueTypeSyncMap -
	KeyTypeValueTypeSyncMap struct {
		m sync.Map
	}
)

// NewKeyTypeValueTypeSyncMap constructs a new SyncMap
func NewKeyTypeValueTypeSyncMap() *KeyTypeValueTypeSyncMap {
	return &KeyTypeValueTypeSyncMap{}
}

// Put -
func (m *KeyTypeValueTypeSyncMap) Put(k KeyType, v *ValueType) {
	m.m.Store(k, v)
}

// Get -
func (m *KeyTypeValueTypeSyncMap) Get(k KeyType) (*ValueType, bool) {
	i, ok := m.m.Load(k)
	if !ok {
		return nil, false
	}

	v, ok := i.(*ValueType)
	if !ok {
		return nil, false
	}

	return v, true
}

// Delete -
func (m *KeyTypeValueTypeSyncMap) Delete(k KeyType) {
	m.m.Delete(k)
}

// Range -
func (m *KeyTypeValueTypeSyncMap) Range(i func(k KeyType, v *ValueType) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return i(k.(KeyType), v.(*ValueType))
	})
}
