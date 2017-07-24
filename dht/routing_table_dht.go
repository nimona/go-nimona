package dht

import "sync"

// RoutingTableSimple ...
type RoutingTableSimple struct {
	mx    sync.RWMutex
	store map[ID]*Peer
}

// Add ...
func (rt *RoutingTableSimple) Add(peer Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	if _, ok := rt.store[peer.ID]; ok {
		return ErrPeerAlreadyExists
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

// Update ...
func (rt *RoutingTableSimple) Update(peer Peer) error {
	rt.mx.Lock()
	defer rt.mx.Unlock()

	if _, ok := rt.store[peer.ID]; !ok {
		return ErrPeerNotFound
	}
	rt.store[peer.ID] = &peer

	return nil
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
