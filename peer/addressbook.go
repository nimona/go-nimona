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
	PutPeerStatus(peerID, address string, status *Status) error
	// Resolve(ctx context.Context, peerID string) (string, error)
	// Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error)
	LoadOrCreateLocalPeerInfo(path string) (*SecretPeerInfo, error)
	CreateNewPeer() (*SecretPeerInfo, error)
	LoadSecretPeerInfo(path string) (*SecretPeerInfo, error)
	StoreSecretPeerInfo(pi *SecretPeerInfo, path string) error
	GetKeyring() *basic.Keyring
}

// Status represents the connection state of a peer
type Status int

const (
	NotConnected Status = iota
	Connected
	CanConnect
	ErrorConnecting
)

// NewAddressBook creates a new addressBook with an empty keyring
func NewAddressBook() AddressBook {
	adb := &addressBook{
		peers:   map[string]*PeerInfo{},
		keyring: basic.NewKeyring(),
	}

	return adb
}

type addressBook struct {
	sync.RWMutex
	peers      map[string]*PeerInfo
	peerStatus map[string]map[string]*Status
	localPeer  *SecretPeerInfo
	keyring    *basic.Keyring
}

func (adb *addressBook) GetKeyring() *basic.Keyring {
	return adb.keyring
}

func (adb *addressBook) PutPeerInfo(peerInfo *PeerInfo) error {
	adb.Lock()
	defer adb.Unlock()
	if adb.localPeer.ID == peerInfo.ID {
		return ErrCannotPutLocalPeerInfo
	}

	if peerInfo.ID == "" {
		return nil
	}

	peerInfo.UpdatedAt = time.Now()

	adb.peers[peerInfo.ID] = peerInfo
	return nil
}

func (adb *addressBook) GetLocalPeerInfo() *SecretPeerInfo {
	return adb.localPeer
}

func (adb *addressBook) PutLocalPeerInfo(peerInfo *SecretPeerInfo) error {
	adb.Lock()
	defer adb.Unlock()
	peerInfo.UpdatedAt = time.Now()
	adb.localPeer = peerInfo
	return nil
}

func (adb *addressBook) PutPeerStatus(peerID, address string,
	status *Status) error {
	adb.Lock()
	defer adb.Unlock()
	adb.peerStatus[peerID][address] = status
	return nil
}

func (adb *addressBook) GetPeerInfo(peerID string) (*PeerInfo, error) {
	adb.RLock()
	defer adb.RUnlock()
	peerInfo, ok := adb.peers[peerID]
	if !ok {
		return nil, ErrNotKnown
	}

	newPeerInfo := &PeerInfo{}
	copier.Copy(newPeerInfo, peerInfo)
	return newPeerInfo, nil
}

func (adb *addressBook) GetAllPeerInfo() ([]*PeerInfo, error) {
	adb.RLock()
	defer adb.RUnlock()
	newPeerInfos := []*PeerInfo{}
	for _, peerInfo := range adb.peers {
		newPeerInfo := &PeerInfo{}
		copier.Copy(newPeerInfo, peerInfo)
		newPeerInfos = append(newPeerInfos, newPeerInfo)
	}
	return newPeerInfos, nil
}
