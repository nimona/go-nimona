package resolver

import (
	"sync"
	"time"

	"nimona.io/pkg/crypto"
)

type (
	peerConnections struct {
		m sync.Map
	}
)

// Get -
func (m *peerConnections) Get(k crypto.PublicKey) *time.Time {
	p, ok := m.m.Load(k)
	if !ok {
		return nil
	}
	t := p.(time.Time)
	return &t
}

// GetOrPut -
func (m *peerConnections) GetOrPut(k crypto.PublicKey) *time.Time {
	p, ok := m.m.LoadOrStore(k, time.Now())
	if !ok {
		return nil
	}
	t := p.(time.Time)
	return &t
}

// Put -
func (m *peerConnections) Put(k crypto.PublicKey) {
	m.m.Store(k, time.Now())
}

// Cleanup -
func (m *peerConnections) Cleanup(d time.Duration) {
	m.m.Range(func(k, v interface{}) bool {
		if v.(time.Time).Add(d).After(time.Now()) {
			m.m.Delete(k)
		}
		return true
	})
}
