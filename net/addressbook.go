package net

import (
	"fmt"
	"sync"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/logger"
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

type peerStatus struct {
	Status         Status
	FailedAttempts int
}

// Status represents the connection state of a peer
type Status string

const (
	// StatusNew when we just received this peer and have not tried connecting
	// to it
	StatusNew Status = "new"
	// StatusConnecting when we have not tried yet connecting to the peer
	StatusConnecting = "connecting"
	// StatusConnected when we are currently connected to this peer
	StatusConnected = "connected"
	// StatusCanConnect when we have connected previously to this peer but
	// disconnected without an error
	StatusCanConnect = "can-connect"
	// StatusError when we failed to connect, or a connection failed with
	// an error
	StatusError = "error"
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
	peerStatus    sync.Map
	localPeerLock sync.RWMutex
	localPeer     *PrivatePeerInfo
}

// PutPeerInfoFromEnvelope stores an envelope with a peer payload
func (ab *AddressBook) PutPeerInfoFromEnvelope(envelope *Envelope) error {
	if ab.localPeer.ID == envelope.Headers.Signer {
		return nil
	}

	ep := envelope.Payload.(PeerInfoPayload)

	// TODO verify envelope?

	exPeer, err := ab.GetPeerInfo(envelope.Headers.Signer)
	if err == nil && exPeer != nil {
		if fmt.Sprintf("%x", exPeer.Envelope.Signature) == fmt.Sprintf("%x", envelope.Signature) {
			return nil
		}
	}

	ab.PutPeerStatus(envelope.Headers.Signer, StatusNew)

	peerInfo := &PeerInfo{
		ID:        envelope.Headers.Signer,
		Addresses: ep.Addresses,
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
// TODO Add timestamps for status changes to help with error checks
// Peer statuses are a bit hacky now:
// * From StatusNew we can only go to StatusConnecting
// * From StatusError we can only go to New and Connected
func (ab *AddressBook) PutPeerStatus(peerID string, status Status) {
	rawStatus, ok := ab.peerStatus.Load(peerID)
	curStatus := StatusNew
	if ok {
		curStatus = rawStatus.(Status)
	}

	if curStatus == StatusError && status == StatusError {
		// TODO too harsh, find another way to remove peers
		ab.peerStatus.Delete(peerID)
		ab.peers.peers.Delete(peerID)
		logger.Info("Removing peer", zap.String("peerID", peerID))
		return
	}

	if curStatus == status {
		return
	}

	if curStatus == StatusNew && status != StatusConnecting {
		// not a valid sequence of events
		return
	}

	if curStatus == StatusError && (status != StatusNew && status != StatusConnected) {
		// from error we cannot go to anything other than new or connected
		// this is a hack until we introduce timestamps for statuses
		return
	}

	logger.Info("Updating peer status", zap.String("curStatus", curStatus), zap.String("newStatus", status))
	ab.peerStatus.Store(peerID, status)
}

// GetPeerStatus returns a peer's status
func (ab *AddressBook) GetPeerStatus(peerID string) Status {
	status, ok := ab.peerStatus.Load(peerID)
	if !ok {
		return StatusConnecting
	}

	return status.(Status)
}
