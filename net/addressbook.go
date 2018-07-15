package net

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/keybase/saltpack/basic"
)

var (
	ErrCannotPutLocalPeerInfo = errors.New("cannot put local peer info")

	peerInfoExpireAfter = time.Hour * 1
)

type PeerManager interface {
	GetLocalPeerInfo() *SecretPeerInfo
	PutLocalPeerInfo(*SecretPeerInfo) error

	GetPeerInfo(peerID string) (*PeerInfo, error)
	GetAllPeerInfo() ([]*PeerInfo, error)
	PutPeerInfoFromMessage(*Message) error

	PutPeerStatus(peerID string, status Status)
	GetPeerStatus(peerID string) Status

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

// NewAddressBook creates a new AddressBook with an empty keyring
func NewAddressBook() *AddressBook {
	adb := &AddressBook{
		identities: &IdentityCollection{},
		peers:      &PeerInfoCollection{},
		keyring:    basic.NewKeyring(),
	}

	return adb
}

type AddressBook struct {
	identities    *IdentityCollection
	peers         *PeerInfoCollection
	peerMessages  sync.Map
	peerStatus    sync.Map
	localPeerLock sync.RWMutex
	localPeer     *SecretPeerInfo
	keyring       *basic.Keyring
}

func (adb *AddressBook) GetKeyring() *basic.Keyring {
	return adb.keyring
}

func (adb *AddressBook) PutPeerInfoFromMessage(message *Message) error {
	if adb.localPeer.ID == message.Headers.Signer {
		return nil
	}

	pip := &PeerInfoPayload{}
	if err := message.DecodePayload(pip); err != nil {
		fmt.Println("ASDFSDF", err)
		return err
	}

	// TODO verify message?

	// TODO check if existing is the same

	// TODO reset connectivity and dates

	// payload, ok := message.Payload.(*PeerInfoPayload)
	// if !ok {
	// 	return errors.New("invalid payload type, expected PeerInfoPayload, got " + reflect.TypeOf(payload).String())
	// }

	peerInfo := &PeerInfo{
		ID:        message.Headers.Signer,
		Addresses: pip.Addresses, // payload.Addresses,
		Message:   message,
	}

	return adb.peers.Put(peerInfo)
}

func (adb *AddressBook) GetLocalPeerInfo() *SecretPeerInfo {
	adb.localPeerLock.RLock()
	newSecretPeerInfo := &SecretPeerInfo{}
	copier.Copy(newSecretPeerInfo, adb.localPeer)
	adb.localPeerLock.RUnlock()
	return newSecretPeerInfo
}

func (adb *AddressBook) PutLocalPeerInfo(peerInfo *SecretPeerInfo) error {
	adb.localPeerLock.Lock()
	newSecretPeerInfo := &SecretPeerInfo{}
	copier.Copy(newSecretPeerInfo, peerInfo)
	adb.localPeer = newSecretPeerInfo
	adb.localPeerLock.Unlock()
	return nil
}

func (adb *AddressBook) GetPeerInfo(peerID string) (*PeerInfo, error) {
	return adb.peers.Get(peerID)
}

func (adb *AddressBook) GetAllPeerInfo() ([]*PeerInfo, error) {
	return adb.peers.All()
}

func (adb *AddressBook) PutPeerStatus(peerID string, status Status) {
	adb.peerStatus.Store(peerID, status)
}

func (adb *AddressBook) GetPeerStatus(peerID string) Status {
	status, ok := adb.peerStatus.Load(peerID)
	if !ok {
		return NotConnected
	}

	return status.(Status)
}
