// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/geoah/genny

package resolver

import (
	"sync"

	"nimona.io/pkg/tilde"
)

type (
	// ChoreHashSyncList -
	ChoreHashSyncList struct {
		m sync.Map
	}
)

// Put -
func (m *ChoreHashSyncList) Put(k tilde.Digest) {
	m.m.Store(k, true)
}

// Exists -
func (m *ChoreHashSyncList) Exists(k tilde.Digest) bool {
	_, ok := m.m.Load(k)
	return ok
}

// Delete -
func (m *ChoreHashSyncList) Delete(k tilde.Digest) {
	m.m.Delete(k)
}

// Range -
func (m *ChoreHashSyncList) Range(i func(k tilde.Digest) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return i(k.(tilde.Digest))
	})
}

// List -
func (m *ChoreHashSyncList) List() []tilde.Digest {
	r := []tilde.Digest{}
	m.m.Range(func(k, v interface{}) bool {
		r = append(r, k.(tilde.Digest))
		return true
	})
	return r
}
