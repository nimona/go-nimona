package mesh

import (
	"crypto/ecdsa"
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
	GetLocalPeerInfo() *PeerInfo
	PutLocalPeerInfo(*PeerInfo) error
	GetPeerInfo(peerID string) (*PeerInfo, error)
	GetAllPeerInfo() ([]*PeerInfo, error)
	PutPeerInfo(*PeerInfo) error
	GetPrivateKey() *ecdsa.PrivateKey
	// Resolve(ctx context.Context, peerID string) (string, error)
	// Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error)
}

func NewRegisty(key *ecdsa.PrivateKey) Registry {
	lp := &PeerInfo{
		ID:        IDFromPublicKey(key.PublicKey),
		Addresses: []string{},
		PublicKey: EncodePublicKey(key.PublicKey),
	}
	lp.Signature, _ = Sign(key, lp.MarshalWithoutSignature())
	reg := &registry{
		localPeer: lp,
		peers:     map[string]*PeerInfo{},
		key:       key,
	}

	return reg
}

type registry struct {
	sync.RWMutex
	peers     map[string]*PeerInfo
	localPeer *PeerInfo
	key       *ecdsa.PrivateKey
}

func (reg *registry) GetPrivateKey() *ecdsa.PrivateKey {
	return reg.key
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

func (reg *registry) GetLocalPeerInfo() *PeerInfo {
	reg.RLock()
	defer reg.RUnlock()
	newPeerInfo := &PeerInfo{}
	copier.Copy(newPeerInfo, reg.localPeer)
	return newPeerInfo
}

func (reg *registry) PutLocalPeerInfo(peerInfo *PeerInfo) error {
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
