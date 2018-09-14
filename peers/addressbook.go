package peers

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"

	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
	"nimona.io/go/log"
)

// AddressBooker provides an interface for mocking our AddressBook
// type AddressBooker interface {
// 	GetPeerInfo(peerID string) (*PeerInfo, error)
// 	GetAllPeerInfo() ([]*PeerInfo, error)
// 	PutPeerInfo(*PeerInfo) error

// 	PutPeerStatus(peerID string, status Status)
// 	GetPeerStatus(peerID string) Status

// 	GetPeerKey() *crypto.Key
// 	GetLocalPeerInfo() (*PeerInfo, error)

// 	AddAddress(addr string) error
// 	RemoveAddress(addr string) error
// 	GetAddresses() []string
// 	AddRelay(relayPeer string) error
// 	GetRelays() []string
// }

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
func NewAddressBook(configPath string) (*AddressBook, error) {
	ab := &AddressBook{
		peers:          &PeerInfoCollection{},
		localAddresses: sync.Map{},
		localRelays:    sync.Map{},
	}

	if err := ab.loadConfig(configPath); err != nil {
		return nil, err
	}

	return ab, nil
}

// AddressBook holds our private peer as well as all known remote peers
type AddressBook struct {
	localKey       *crypto.Key
	localAddresses sync.Map
	localRelays    sync.Map
	peers          *PeerInfoCollection
	peerStatus     sync.Map
}

// GetPeerKey returns the local peer's key
func (ab *AddressBook) GetPeerKey() *crypto.Key {
	return ab.localKey
}

// PutPeerInfo stores an block with a peer payload
func (ab *AddressBook) PutPeerInfo(peerInfo *PeerInfo) error {
	if len(peerInfo.Addresses) == 0 {
		return errors.New("missing addresses")
	}

	if peerInfo.Thumbprint() == ab.GetLocalPeerInfo().Thumbprint() {
		return nil
	}

	peerThumbprint := peerInfo.Thumbprint()
	exPeer, err := ab.GetPeerInfo(peerThumbprint)
	if err == nil && exPeer != nil {
		if fmt.Sprintf("%x", exPeer.Signature) == fmt.Sprintf("%x", peerInfo.Signature) {
			return nil
		}
	}

	ab.PutPeerStatus(peerThumbprint, StatusNew)
	return ab.peers.Put(peerInfo)
}

// GetLocalPeerInfo returns our private peer info
func (ab *AddressBook) GetLocalPeerInfo() *PeerInfo {
	addresses := ab.GetLocalAddresses()
	addresses = append(addresses, ab.GetLocalRelays()...)

	pi := &PeerInfo{
		Addresses: addresses,
	}

	sig, err := blocks.Sign(pi, ab.GetPeerKey())
	if err != nil {
		panic(err)
	}

	pi.Signature = sig
	return pi
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
		log.DefaultLogger.Info("Removing peer", zap.String("peerID", peerID))
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

	log.DefaultLogger.Info("Updating peer status",
		zap.String("curStatus", string(curStatus)),
		zap.String("newStatus", string(status)))
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

// AddAddress for local peer
func (ab *AddressBook) AddAddress(addresses ...string) error {
	for _, address := range addresses {
		ab.localAddresses.Store(address, true)
	}
	return nil
}

// RemoveAddress for local peer
func (ab *AddressBook) RemoveAddress(addr string) error {
	ab.localAddresses.Delete(addr)
	return nil
}

// GetLocalAddresses for local peer
func (ab *AddressBook) GetLocalAddresses() []string {
	addresses := []string{}
	ab.localAddresses.Range(func(key, value interface{}) bool {
		addresses = append(addresses, key.(string))
		return true
	})
	return addresses
}

// AddRelay for local peer
func (ab *AddressBook) AddRelay(relayPeers ...string) error {
	for _, relayPeer := range relayPeers {
		relayPeer = strings.Replace(relayPeer, "relay:", "", 1)
		relayPeer = "relay:" + relayPeer
		ab.localRelays.Store(relayPeer, true)
	}
	return nil
}

// GetLocalRelays for peer
func (ab *AddressBook) GetLocalRelays() []string {
	relays := []string{}
	ab.localRelays.Range(func(key, value interface{}) bool {
		relays = append(relays, key.(string))
		return true
	})

	return relays
}
