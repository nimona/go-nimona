package resolver

import (
	"fmt"
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

type (
	peerCache struct {
		m sync.Map
	}
)

// Put -
func (m *peerCache) Put(p *peer.Peer) {
	m.m.Store(p.PublicKey(), p)
}

// Get -
func (m *peerCache) Get(k crypto.PublicKey) (*peer.Peer, error) {
	p, ok := m.m.Load(k)
	if !ok {
		return nil, fmt.Errorf("missing")
	}
	return p.(*peer.Peer), nil
}

// List -
func (m *peerCache) List() []*peer.Peer {
	ps := []*peer.Peer{}
	m.m.Range(func(_, p interface{}) bool {
		ps = append(ps, p.(*peer.Peer))
		return true
	})
	return ps
}
