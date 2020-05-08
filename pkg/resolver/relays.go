package resolver

import (
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

type (
	relays struct {
		m sync.Map
	}
)

// Put -
func (m *relays) Put(v *peer.Peer) {
	m.m.Store(v.PublicKey(), v)
}

// Delete -
func (m *relays) Delete(k crypto.PublicKey) {
	m.m.Delete(k)
}

// List -
func (m *relays) List() []*peer.Peer {
	hs := []*peer.Peer{}
	m.m.Range(func(k, v interface{}) bool {
		hs = append(hs, v.(*peer.Peer))
		return true
	})
	return hs
}
