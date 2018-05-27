package mesh

import (
	"errors"
	"sync"
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
	GetLocalPeerInfo() *SecretPeerInfo
	PutLocalPeerInfo(*SecretPeerInfo) error // TODO Deprecate
	GetPeerInfo(peerID string) (*PeerInfo, error)
	GetAllPeerInfo() ([]*PeerInfo, error)
	PutPeerInfo(*PeerInfo) error
	// Resolve(ctx context.Context, peerID string) (string, error)
	// Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error)
}

func NewRegisty(lp *SecretPeerInfo) Registry {
	reg := &registry{
		localPeer: lp,
		peers:     map[string]*PeerInfo{},
	}

	return reg
}

type registry struct {
	sync.RWMutex
	peers     map[string]*PeerInfo
	localPeer *SecretPeerInfo
}

func (reg *registry) PutPeerInfo(peerInfo *PeerInfo) error {
	reg.Lock()
	defer reg.Unlock()
	if reg.localPeer.ID == peerInfo.ID {
		return ErrCannotPutLocalPeerInfo
	}

	if peerInfo.ID == "" || len(peerInfo.Addresses) == 0 {
		return nil
	}

	reg.peers[peerInfo.ID] = peerInfo
	return nil
}

func (reg *registry) GetLocalPeerInfo() *SecretPeerInfo {
	return reg.localPeer
}

func (reg *registry) PutLocalPeerInfo(peerInfo *SecretPeerInfo) error {
	reg.Lock()
	defer reg.Unlock()
	reg.localPeer = peerInfo
	return nil
}

func (reg *registry) GetPeerInfo(peerID string) (*PeerInfo, error) {
	reg.RLock()
	defer reg.RUnlock()
	peerInfo, ok := reg.peers[peerID]
	if !ok {
		return nil, ErrNotKnown
	}

	newPeerInfo := &PeerInfo{}
	copier.Copy(newPeerInfo, peerInfo)
	return newPeerInfo, nil
}

func (reg *registry) GetAllPeerInfo() ([]*PeerInfo, error) {
	reg.RLock()
	defer reg.RUnlock()
	newPeerInfos := []*PeerInfo{}
	for _, peerInfo := range reg.peers {
		newPeerInfo := &PeerInfo{}
		copier.Copy(newPeerInfo, peerInfo)
		newPeerInfos = append(newPeerInfos, newPeerInfo)
	}
	return newPeerInfos, nil
}
