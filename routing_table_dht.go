package dht

import (
	"sync"

	net "github.com/nimona/go-nimona-net"
)

// RoutingTableSimple ...
type RoutingTableSimple struct {
	mx    sync.RWMutex
	store map[string]net.Peer
}

// Save ...
func (rt *RoutingTableSimple) Save(peer net.Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	// If peer exists update address
	if _, ok := rt.store[peer.ID]; ok {
		rt.store[peer.ID].Address = peer.Address
	}
	rt.store[peer.ID] = peer

	return nil
}

// Remove ...
func (rt *RoutingTableSimple) Remove(peer net.Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	if _, ok := rt.store[peer.ID]; !ok {
		return ErrPeerNotFound
	}
	delete(rt.store, peer.ID)

	return nil
}

// Get ...
func (rt *RoutingTableSimple) Get(string string) (net.Peer, error) {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	pr, ok := rt.store[string]
	if !ok {
		return net.Peer{}, ErrPeerNotFound
	}

	return pr, nil
}

func (rt *RoutingTableSimple) GetPeerIDs() ([]string, error) {
	rt.mx.Lock()
	defer rt.mx.Unlock()
	ids := make([]string, len(rt.store))
	i := 0
	for _, peer := range rt.store {
		ids[i] = peer.ID
		i++
	}
	return ids, nil
}

func NewSimpleRoutingTable() *RoutingTableSimple {
	return &RoutingTableSimple{
		store: make(map[string]net.Peer),
	}
}
