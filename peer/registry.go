package peer

import (
	"errors"
	"sync"
	"time"

	"github.com/keybase/saltpack/basic"

	"github.com/jinzhu/copier"
)

var (
	ErrNotKnown               = errors.New("not found")
	ErrCannotPutLocalPeerInfo = errors.New("cannot put local peer info")
)

var (
	peerInfoExpireAfter = time.Hour * 1
)

type AddressBook interface {
	GetLocalPeerInfo() *SecretPeerInfo
	PutLocalPeerInfo(*SecretPeerInfo) error // TODO Deprecate
	GetPeerInfo(peerID string) (*PeerInfo, error)
	GetAllPeerInfo() ([]*PeerInfo, error)
	PutPeerInfo(*PeerInfo) error
	// Resolve(ctx context.Context, peerID string) (string, error)
	// Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error)
	LoadOrCreateLocalPeerInfo(path string) (*SecretPeerInfo, error)
	CreateNewPeer() (*SecretPeerInfo, error)
	LoadSecretPeerInfo(path string) (*SecretPeerInfo, error)
	StoreSecretPeerInfo(pi *SecretPeerInfo, path string) error
	GetKeyring() *basic.Keyring
}

// NewAddressBook creates a new addressBook with an empty keyring
func NewAddressBook() AddressBook {
	reg := &addressBook{
		peers:   map[string]*PeerInfo{},
		keyring: basic.NewKeyring(),
	}

	return reg
}

type addressBook struct {
	sync.RWMutex
	peers     map[string]*PeerInfo
	localPeer *SecretPeerInfo
	keyring   *basic.Keyring
}

func (reg *addressBook) GetKeyring() *basic.Keyring {
	return reg.keyring
}

func (reg *addressBook) PutPeerInfo(peerInfo *PeerInfo) error {
	reg.Lock()
	defer reg.Unlock()
	if reg.localPeer.ID == peerInfo.ID {
		return ErrCannotPutLocalPeerInfo
	}

	if peerInfo.ID == "" {
		return nil
	}

	peerInfo.UpdatedAt = time.Now()

	reg.peers[peerInfo.ID] = peerInfo
	return nil
}

func (reg *addressBook) GetLocalPeerInfo() *SecretPeerInfo {
	return reg.localPeer
}

func (reg *addressBook) PutLocalPeerInfo(peerInfo *SecretPeerInfo) error {
	reg.Lock()
	defer reg.Unlock()
	peerInfo.UpdatedAt = time.Now()
	reg.localPeer = peerInfo
	return nil
}

func (reg *addressBook) GetPeerInfo(peerID string) (*PeerInfo, error) {
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

func (reg *addressBook) GetAllPeerInfo() ([]*PeerInfo, error) {
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
