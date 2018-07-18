package net

import (
	"errors"
	"sync"
	"time"

	"github.com/jinzhu/copier"
)

var (
	ErrCannotPutLocalPeerInfo = errors.New("cannot put local peer info")

	peerInfoExpireAfter = time.Hour * 1
)

type PeerManager interface {
	GetLocalPeerInfo() *PrivatePeerInfo
	PutLocalPeerInfo(*PrivatePeerInfo) error

	GetPeerInfo(peerID string) (*PeerInfo, error)
	GetAllPeerInfo() ([]*PeerInfo, error)
	PutPeerInfoFromEnvelope(*Envelope) error

	PutPeerStatus(peerID string, status Status)
	GetPeerStatus(peerID string) Status

	// Resolve(ctx context.Context, peerID string) (string, error)
	// Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error)
	LoadOrCreateLocalPeerInfo(path string) (*PrivatePeerInfo, error)
	CreateNewPeer() (*PrivatePeerInfo, error)
	LoadPrivatePeerInfo(path string) (*PrivatePeerInfo, error)
	StorePrivatePeerInfo(pi *PrivatePeerInfo, path string) error
}

// Status represents the connection state of a peer
type Status int

const (
	NotConnected Status = iota
	Connected
	CanConnect
	ErrorConnecting
)

// NewAddressBook creates a new AddressBook
func NewAddressBook() *AddressBook {
	ab := &AddressBook{
		identities: &IdentityCollection{},
		peers:      &PeerInfoCollection{},
	}

	return ab
}

type AddressBook struct {
	identities    *IdentityCollection
	peers         *PeerInfoCollection
	peerEnvelopes sync.Map
	peerStatus    sync.Map
	localPeerLock sync.RWMutex
	localPeer     *PrivatePeerInfo
}

func (ab *AddressBook) PutPeerInfoFromEnvelope(envelope *Envelope) error {
	if ab.localPeer.ID == envelope.Headers.Signer {
		return nil
	}

	pip := envelope.Payload.(PeerInfoPayload)

	// TODO verify envelope?

	// TODO check if existing is the same

	// TODO reset connectivity and dates

	// payload, ok := envelope.Payload.(PeerInfoPayload)
	// if !ok {
	// 	return errors.New("invalid payload type, expected PeerInfoPayload, got " + reflect.TypeOf(payload).String())
	// }

	peerInfo := &PeerInfo{
		ID:        envelope.Headers.Signer,
		Addresses: pip.Addresses, // payload.Addresses,
		Envelope:  envelope,
	}

	return ab.peers.Put(peerInfo)
}

func (ab *AddressBook) GetLocalPeerInfo() *PrivatePeerInfo {
	ab.localPeerLock.RLock()
	newPrivatePeerInfo := &PrivatePeerInfo{}
	copier.Copy(newPrivatePeerInfo, ab.localPeer)
	ab.localPeerLock.RUnlock()
	return newPrivatePeerInfo
}

func (ab *AddressBook) PutLocalPeerInfo(peerInfo *PrivatePeerInfo) error {
	ab.localPeerLock.Lock()
	newPrivatePeerInfo := &PrivatePeerInfo{}
	copier.Copy(newPrivatePeerInfo, peerInfo)
	ab.localPeer = newPrivatePeerInfo
	ab.localPeerLock.Unlock()
	return nil
}

func (ab *AddressBook) GetPeerInfo(peerID string) (*PeerInfo, error) {
	return ab.peers.Get(peerID)
}

func (ab *AddressBook) GetAllPeerInfo() ([]*PeerInfo, error) {
	return ab.peers.All()
}

func (ab *AddressBook) PutPeerStatus(peerID string, status Status) {
	ab.peerStatus.Store(peerID, status)
}

func (ab *AddressBook) GetPeerStatus(peerID string) Status {
	status, ok := ab.peerStatus.Load(peerID)
	if !ok {
		return NotConnected
	}

	return status.(Status)
}
