package peers

import (
	"strings"
	"sync"

	"go.uber.org/zap"

	"nimona.io/go/crypto"
	"nimona.io/go/log"
)

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
	LocalHostname  string
	localAddresses sync.Map
	localRelays    sync.Map
	peers          *PeerInfoCollection
	peerStatus     sync.Map
	aliases        sync.Map
}

// GetLocalPeerKey returns the local peer's key
// TODO make this an attribute, is there any reason for this to be a method?
func (ab *AddressBook) GetLocalPeerKey() *crypto.Key {
	return ab.localKey
}

// HandleObject of any type
// func (ab *AddressBook) HandleObject(o *encoding.Object) error {
// 	switch o.GetType() {
// 	case "nimona.io/peer.info":
// 		v, err := NewPeerInfoFromObject(o)
// 		if err != nil {
// 			return err
// 		}
// 		if err := ab.PutPeerInfo(v); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// PutPeerInfo stores an block with a peer payload
func (ab *AddressBook) PutPeerInfo(peerInfo *PeerInfo) error {
	if peerInfo.Thumbprint() == ab.GetLocalPeerInfo().Thumbprint() {
		return nil
	}

	peerThumbprint := peerInfo.Thumbprint()
	// exPeer, err := ab.GetPeerInfo(peerThumbprint)
	// if err == nil && exPeer != nil {
	// 	if fmt.Sprintf("%x", exPeer.Signature) == fmt.Sprintf("%x", peerInfo.Signature) {
	// 		return nil
	// 	}
	// }

	ab.PutPeerStatus(peerThumbprint, StatusNew)
	return ab.peers.Put(peerInfo)
}

// GetLocalPeerInfo returns our private peer info
// TODO make this an attribute, is there any reason for this to be a method?
func (ab *AddressBook) GetLocalPeerInfo() *PeerInfo {
	addresses := ab.GetLocalPeerAddresses()

	p := &PeerInfo{
		Addresses: addresses,
	}

	spo, err := crypto.Sign(p.ToObject(), ab.GetLocalPeerKey())
	if err != nil {
		panic(err)
	}

	if err := p.FromObject(spo); err != nil {
		panic(err)
	}

	return p
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

// AddLocalPeerAddress for local peer
func (ab *AddressBook) AddLocalPeerAddress(addresses ...string) error {
	for _, address := range addresses {
		ab.localAddresses.Store(address, true)
	}
	return nil
}

// RemoveLocalPeerAddress for local peer
func (ab *AddressBook) RemoveLocalPeerAddress(addr string) error {
	ab.localAddresses.Delete(addr)
	return nil
}

// GetLocalPeerAddresses for local peer
func (ab *AddressBook) GetLocalPeerAddresses() []string {
	addresses := []string{}
	ab.localAddresses.Range(func(key, value interface{}) bool {
		addresses = append(addresses, key.(string))
		return true
	})
	addresses = append(addresses, ab.GetLocalPeerRelays()...)
	return addresses
}

// AddLocalPeerRelay for local peer
func (ab *AddressBook) AddLocalPeerRelay(relayPeers ...string) error {
	for _, relayPeer := range relayPeers {
		relayPeer = strings.Replace(relayPeer, "relay:", "", 1)
		relayPeer = "relay:" + relayPeer
		ab.localRelays.Store(relayPeer, true)
	}
	return nil
}

// GetLocalPeerRelays for peer
func (ab *AddressBook) GetLocalPeerRelays() []string {
	relays := []string{}
	ab.localRelays.Range(func(key, value interface{}) bool {
		relays = append(relays, key.(string))
		return true
	})

	return relays
}

func (ab *AddressBook) SetAlias(k *crypto.Key, v string) {
	ab.aliases.Store(k, v)
}

func (ab *AddressBook) GetAlias(k *crypto.Key) string {
	v, ok := ab.aliases.Load(k)
	if ok {
		return v.(string)
	}
	return k.HashBase58()
}
