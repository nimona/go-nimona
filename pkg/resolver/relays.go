package resolver

import (
	"sync"

	"nimona.io/pkg/crypto"
)

type (
	relays struct {
		m sync.Map
	}
)

// Put -
func (m *relays) Put(k crypto.PublicKey) {
	m.m.Store(k, true)
}

// Delete -
func (m *relays) Delete(k crypto.PublicKey) {
	m.m.Delete(k)
}

// List -
func (m *relays) List() []crypto.PublicKey {
	hs := []crypto.PublicKey{}
	m.m.Range(func(k, v interface{}) bool {
		hs = append(hs, k.(crypto.PublicKey))
		return true
	})
	return hs
}
