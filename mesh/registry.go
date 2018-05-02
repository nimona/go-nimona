package mesh

import (
	"errors"
	"time"

	"github.com/jinzhu/copier"
)

var (
	ErrNotKnown               = errors.New("not found")
	ErrCannotPutLocalPeerInfo = errors.New("cannot put local peer info")
)

var (
	peerInfoExpireAfter = time.Hour * 1
)

type Registry interface {
	GetLocalPeerInfo() *PeerInfo
	PutLocalPeerInfo(*PeerInfo) error
	GetPeerInfo(peerID string) (*PeerInfo, error)
	GetAllPeerInfo() ([]*PeerInfo, error)
	PutPeerInfo(*PeerInfo) error
	// Resolve(ctx context.Context, peerID string) (string, error)
	// Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error)
}

func NewRegisty(peerID string) (Registry, error) {
	reg := &registry{
		localPeer: &PeerInfo{
			ID:        peerID,
			Protocols: map[string][]string{},
		},
		peers: map[string]*PeerInfo{},
	}

	return reg, nil
}

type registry struct {
	peers     map[string]*PeerInfo
	localPeer *PeerInfo
}

func (reg *registry) PutPeerInfo(peerInfo *PeerInfo) error {
	if reg.localPeer.ID == peerInfo.ID {
		return ErrCannotPutLocalPeerInfo
	}

	reg.peers[peerInfo.ID] = peerInfo
	return nil
}

func (reg *registry) GetLocalPeerInfo() *PeerInfo {
	newPeerInfo := &PeerInfo{}
	copier.Copy(newPeerInfo, reg.localPeer)
	return newPeerInfo
}

func (reg *registry) PutLocalPeerInfo(peerInfo *PeerInfo) error {
	reg.localPeer = peerInfo
	return nil
}

func (reg *registry) GetPeerInfo(peerID string) (*PeerInfo, error) {
	peerInfo, ok := reg.peers[peerID]
	if !ok {
		return nil, ErrNotKnown
	}

	newPeerInfo := &PeerInfo{}
	copier.Copy(newPeerInfo, peerInfo)
	return newPeerInfo, nil
}

func (reg *registry) GetAllPeerInfo() ([]*PeerInfo, error) {
	newPeerInfos := []*PeerInfo{}
	for _, peerInfo := range reg.peers {
		newPeerInfo := &PeerInfo{}
		copier.Copy(newPeerInfo, peerInfo)
		newPeerInfos = append(newPeerInfos, newPeerInfo)
	}
	return newPeerInfos, nil
}
