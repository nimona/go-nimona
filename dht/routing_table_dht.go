package dht

import "sync"

// RoutingTableSimple ...
type RoutingTableSimple struct {
	mx    sync.RWMutex
	store map[ID]*Peer
}

// Save ...
func (rt *RoutingTableSimple) Save(peer Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	// If peer exists update address
	if _, ok := rt.store[peer.ID]; ok {
		rt.store[peer.ID].Address = peer.Address
	}
	rt.store[peer.ID] = &peer

	return nil
}

// Remove ...
func (rt *RoutingTableSimple) Remove(peer Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	if _, ok := rt.store[peer.ID]; !ok {
		return ErrPeerNotFound
	}
	delete(rt.store, peer.ID)

	return nil
}

// Get ...
func (rt *RoutingTableSimple) Get(id ID) (Peer, error) {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	pr, ok := rt.store[id]
	if !ok {
		return Peer{}, ErrPeerNotFound
	}

	return *pr, nil
}

func (rt *RoutingTableSimple) GetPeerIDs() ([]ID, error) {
	rt.mx.Lock()
	defer rt.mx.Unlock()
	ids := make([]ID, len(rt.store))
	i := 0
	for _, peer := range rt.store {
		ids[i] = peer.ID
		i++
	}
	return ids, nil
}

func NewSimpleRoutingTable() *RoutingTableSimple {
	return &RoutingTableSimple{
		store: make(map[ID]*Peer),
	}
}
