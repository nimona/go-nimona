package net

import (
	"sync"
)

type IdentityCollection struct {
	peers sync.Map
}

func (c *IdentityCollection) All() ([]*Identity, error) {
	peers := []*Identity{}
	c.peers.Range(func(k, v interface{}) bool {
		peers = append(peers, v.(*Identity))
		return true
	})

	return peers, nil
}

func (c *IdentityCollection) Get(peerID string) (*Identity, error) {
	peer, ok := c.peers.Load(peerID)
	if !ok || peer == nil {
		return nil, ErrNotFound
	}

	return peer.(*Identity), nil
}

func (c *IdentityCollection) Put(peer *Identity) error {
	c.peers.Store(peer.ID, peer)
	return nil
}
