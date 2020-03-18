package syncmap

import (
	"sync"

	"github.com/geoah/genny/generic"
)

type (
	KeyType     generic.Type // nolint
	ValueType   generic.Type // nolint
	SyncmapName string       // nolint
	// SyncmapNameMap -
	SyncmapNameMap struct {
		m sync.Map
	}
)

// NewSyncmapNameMap constructs a new SyncMap
func NewSyncmapNameMap() *SyncmapNameMap {
	return &SyncmapNameMap{}
}

// GetOrPut -
func (m *SyncmapNameMap) GetOrPut(k KeyType, v *ValueType) (*ValueType, bool) {
	nv, ok := m.m.LoadOrStore(k, v)
	return nv.(*ValueType), ok
}

// Put -
func (m *SyncmapNameMap) Put(k KeyType, v *ValueType) {
	m.m.Store(k, v)
}

// Get -
func (m *SyncmapNameMap) Get(k KeyType) (*ValueType, bool) {
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
func (m *SyncmapNameMap) Delete(k KeyType) {
	m.m.Delete(k)
}

// Range -
func (m *SyncmapNameMap) Range(i func(k KeyType, v *ValueType) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return i(k.(KeyType), v.(*ValueType))
	})
}
