// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/geoah/genny

package resolver

import (
	"sync"

	"nimona.io/pkg/object"
)

type (
	// ObjectCIDSyncList -
	ObjectCIDSyncList struct {
		m sync.Map
	}
)

// Put -
func (m *ObjectCIDSyncList) Put(k object.CID) {
	m.m.Store(k, true)
}

// Exists -
func (m *ObjectCIDSyncList) Exists(k object.CID) bool {
	_, ok := m.m.Load(k)
	return ok
}

// Delete -
func (m *ObjectCIDSyncList) Delete(k object.CID) {
	m.m.Delete(k)
}

// Range -
func (m *ObjectCIDSyncList) Range(i func(k object.CID) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return i(k.(object.CID))
	})
}

// List -
func (m *ObjectCIDSyncList) List() []object.CID {
	r := []object.CID{}
	m.m.Range(func(k, v interface{}) bool {
		r = append(r, k.(object.CID))
		return true
	})
	return r
}
