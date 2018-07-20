package net

import (
	"sync"

	"github.com/jinzhu/copier"
)

// PeerManager provides an interface for mocking our AddressBook
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
	// NotConnected when we have not tried yet connecting to the peer
	NotConnected Status = iota
	// Connected when we are currently connected to this peer
	Connected
	// CanConnect when we have connected previously to this peer but
	// disconnected without an error
	CanConnect
	// ErrorConnecting when we failed to connect, or a connection failed with
	// an error
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

// AddressBook holds our private peer as well as all known remote peers
type AddressBook struct {
	identities    *IdentityCollection
	peers         *PeerInfoCollection
	peerEnvelopes sync.Map
	peerStatus    sync.Map
	localPeerLock sync.RWMutex
	localPeer     *PrivatePeerInfo
}

// PutPeerInfoFromEnvelope stores an envelope with a peer payload
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

// GetLocalPeerInfo returns our private peer info
func (ab *AddressBook) GetLocalPeerInfo() *PrivatePeerInfo {
	ab.localPeerLock.RLock()
	newPrivatePeerInfo := &PrivatePeerInfo{}
	copier.Copy(newPrivatePeerInfo, ab.localPeer)
	ab.localPeerLock.RUnlock()
	return newPrivatePeerInfo
}

// PutLocalPeerInfo puts our local peer info
func (ab *AddressBook) PutLocalPeerInfo(peerInfo *PrivatePeerInfo) error {
	ab.localPeerLock.Lock()
	newPrivatePeerInfo := &PrivatePeerInfo{}
	copier.Copy(newPrivatePeerInfo, peerInfo)
	ab.localPeer = newPrivatePeerInfo
	ab.localPeerLock.Unlock()
	return nil
}

// GetPeerInfo returns a peer info from its id
func (ab *AddressBook) GetPeerInfo(peerID string) (*PeerInfo, error) {
	return ab.peers.Get(peerID)
}

// GetAllPeerInfo returns all know peer infos
func (ab *AddressBook) GetAllPeerInfo() ([]*PeerInfo, error) {
	return ab.peers.All()
}

// PutPeerStatus updates a peer's status
func (ab *AddressBook) PutPeerStatus(peerID string, status Status) {
	ab.peerStatus.Store(peerID, status)
}

// GetPeerStatus returns a peer's status
func (ab *AddressBook) GetPeerStatus(peerID string) Status {
	status, ok := ab.peerStatus.Load(peerID)
	if !ok {
		return NotConnected
	}

	return status.(Status)
}
