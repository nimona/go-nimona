package peers

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

// PeerInfoCollection allows concurrent access to peerinfos
type PeerInfoCollection struct {
	peers sync.Map
}

// All returns all items in the collection
func (c *PeerInfoCollection) All() ([]*PeerInfo, error) {
	peers := []*PeerInfo{}
	c.peers.Range(func(k, v interface{}) bool {
		newPeerInfo := &PeerInfo{}
		copier.Copy(newPeerInfo, v.(*PeerInfo))
		peers = append(peers, newPeerInfo)
		return true
	})

	return peers, nil
}

// Get retuns a single item from the collection given its id
func (c *PeerInfoCollection) Get(peerID string) (*PeerInfo, error) {
	peerInfo, ok := c.peers.Load(peerID)
	if !ok || peerInfo == nil {
		return nil, ErrNotFound
	}

	newPeerInfo := &PeerInfo{}
	copier.Copy(newPeerInfo, peerInfo.(*PeerInfo))
	return newPeerInfo, nil
}

// Put adds or overwrites an item in the collection
func (c *PeerInfoCollection) Put(peerInfo *PeerInfo) error {
	c.peers.Store(peerInfo.ID, peerInfo)
	return nil
}
