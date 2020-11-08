package peerstore

import (
	"fmt"
	"sync"
	"time"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/peer"
)

type (
	PeerCache struct {
		m sync.Map
	}
)

type entry struct {
	ttl       time.Duration
	createdAt time.Time
	pr        *peer.Peer
}

func NewPeerCache(gcTime time.Duration) *PeerCache {
	pc := &PeerCache{
		m: sync.Map{},
	}
	go func() {
		for {
			time.Sleep(gcTime)
			pc.m.Range(func(key, value interface{}) bool {
				e := value.(entry)
				if e.ttl != 0 {
					now := time.Now()
					diff := now.Sub(e.createdAt)
					if diff >= e.ttl {
						pc.m.Delete(key)
					}
				}
				return true
			})
		}
	}()
	return pc
}

// Put -
func (m *PeerCache) Put(p *peer.Peer, ttl time.Duration) {
	m.m.Store(p.PublicKey(), entry{
		ttl:       ttl,
		createdAt: time.Now(),
		pr:        p,
	})
}

// Put -
func (m *PeerCache) Touch(k crypto.PublicKey, ttl time.Duration) {
	v, ok := m.m.Load(k)
	if !ok {
		return
	}
	e := v.(entry)
	m.m.Store(k, entry{
		ttl:       ttl,
		createdAt: time.Now(),
		pr:        e.pr,
	})
}

// Get -
func (m *PeerCache) Get(k crypto.PublicKey) (*peer.Peer, error) {
	p, ok := m.m.Load(k)
	if !ok {
		return nil, fmt.Errorf("missing")
	}
	return p.(entry).pr, nil
}

// Remove -
func (m *PeerCache) Remove(k crypto.PublicKey) {
	m.m.Delete(k)
}

// List -
func (m *PeerCache) List() []*peer.Peer {
	ps := []*peer.Peer{}
	m.m.Range(func(_, p interface{}) bool {
		ps = append(ps, p.(entry).pr)
		return true
	})
	return ps
}

// Lookup -
func (m *PeerCache) Lookup(q hyperspace.Bloom) []*peer.Peer {
	ps := []*peer.Peer{}
	m.m.Range(func(_, p interface{}) bool {
		if hyperspace.Bloom(p.(entry).pr.QueryVector).Test(q) {
			ps = append(ps, p.(entry).pr)
		}
		return true
	})
	return ps
}
