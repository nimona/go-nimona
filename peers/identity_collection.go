package peers

import (
	"sync"
)

// IdentityCollection allows concurrent access to peerinfos
type IdentityCollection struct {
	peers sync.Map
}

// All returns all items in the collection
func (c *IdentityCollection) All() ([]*Identity, error) {
	peers := []*Identity{}
	c.peers.Range(func(k, v interface{}) bool {
		peers = append(peers, v.(*Identity))
		return true
	})

	return peers, nil
}

// Get retuns a single item from the collection given its id
func (c *IdentityCollection) Get(peerID string) (*Identity, error) {
	peer, ok := c.peers.Load(peerID)
	if !ok || peer == nil {
		return nil, ErrNotFound
	}

	return peer.(*Identity), nil
}

// Put adds or overwrites an item in the collection
func (c *IdentityCollection) Put(peer *Identity) error {
	c.peers.Store(peer.ID, peer)
	return nil
}
