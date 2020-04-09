// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/geoah/genny

package connmanager

import (
	"sync"

	"nimona.io/pkg/crypto"
)

type (
	connections string // nolint
	// ConnectionsMap -
	ConnectionsMap struct {
		m sync.Map
	}
)

// NewConnectionsMap constructs a new SyncMap
func NewConnectionsMap() *ConnectionsMap {
	return &ConnectionsMap{}
}

// GetOrPut -
func (m *ConnectionsMap) GetOrPut(k crypto.PublicKey, v *peerbox) (*peerbox, bool) {
	nv, ok := m.m.LoadOrStore(k, v)
	return nv.(*peerbox), ok
}

// Put -
func (m *ConnectionsMap) Put(k crypto.PublicKey, v *peerbox) {
	m.m.Store(k, v)
}

// Get -
func (m *ConnectionsMap) Get(k crypto.PublicKey) (*peerbox, bool) {
	i, ok := m.m.Load(k)
	if !ok {
		return nil, false
	}

	v, ok := i.(*peerbox)
	if !ok {
		return nil, false
	}

	return v, true
}

// Delete -
func (m *ConnectionsMap) Delete(k crypto.PublicKey) {
	m.m.Delete(k)
}

// Range -
func (m *ConnectionsMap) Range(i func(k crypto.PublicKey, v *peerbox) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return i(k.(crypto.PublicKey), v.(*peerbox))
	})
}
