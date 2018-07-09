package peer

import (
	"errors"
	"sync"

	"github.com/jinzhu/copier"
)

var (
	ErrNotFound = errors.New("not found")
)

type PeerInfoCollection struct {
	peers sync.Map
}

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

func (c *PeerInfoCollection) Get(peerID string) (*PeerInfo, error) {
	peerInfo, ok := c.peers.Load(peerID)
	if !ok || peerInfo == nil {
		return nil, ErrNotFound
	}

	newPeerInfo := &PeerInfo{}
	copier.Copy(newPeerInfo, peerInfo.(*PeerInfo))
	return newPeerInfo, nil
}

func (c *PeerInfoCollection) Put(peerInfo *PeerInfo) error {
	newPeerInfo := &PeerInfo{}
	copier.Copy(newPeerInfo, peerInfo)
	c.peers.Store(newPeerInfo.ID, newPeerInfo)
	return nil
}
