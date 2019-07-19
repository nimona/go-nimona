package peer

import (
	"errors"
	"sync"

	"github.com/jinzhu/copier"
)

var (
	// ErrNotFound is returned wheh a requqested item in the collection does
	// not exist
	ErrNotFound = errors.New("peer not found")
)

// PeerCollection allows concurrent access to peers
type PeerCollection struct {
	peers sync.Map
}

// All returns all items in the collection
func (c *PeerCollection) All() ([]*Peer, error) {
	peers := []*Peer{}
	c.peers.Range(func(k, v interface{}) bool {
		newPeer := &Peer{}
		copier.Copy(newPeer, v.(*Peer)) // nolint: errcheck
		peers = append(peers, newPeer)
		return true
	})

	return peers, nil
}

// Get retuns a single item from the collection given its id
func (c *PeerCollection) Get(fingerprint string) (*Peer, error) {
	peer, ok := c.peers.Load(fingerprint)
	if !ok || peer == nil {
		return nil, ErrNotFound
	}

	newPeer := &Peer{}
	if err := copier.Copy(&newPeer, peer.(*Peer)); err != nil {
		return nil, err
	}

	return newPeer, nil
}

// Put adds or overwrites an item in the collection
func (c *PeerCollection) Put(peer *Peer) error {
	c.peers.Store(peer.Signature.PublicKey.Fingerprint(), peer)
	return nil
}
