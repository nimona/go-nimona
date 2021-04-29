package localpeer

import (
	"sort"
	"sync"

	"nimona.io/internal/rand"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

//go:generate mockgen -destination=../localpeermock/localpeermock_generated.go -package=localpeermock -source=localpeer.go
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=cids_generated.go -imp=nimona.io/pkg/object -pkg=localpeer gen "KeyType=object.CID"
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=certificates_generated.go -imp=nimona.io/pkg/peer -pkg=localpeer gen "KeyType=*object.Certificate"
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=addresses_generated.go -imp=nimona.io/pkg/peer -pkg=localpeer gen "KeyType=string"

type (
	LocalPeer interface {
		// TODO merge peer/id methods, use .Usage to distinguish
		GetPeerKey() crypto.PrivateKey
		SetPeerKey(crypto.PrivateKey)
		GetCIDs() []object.CID
		RegisterCIDs(...object.CID)
		GetContentTypes() []string
		RegisterContentTypes(...string)
		GetAddresses() []string
		RegisterAddresses(...string)
		GetRelays() []*peer.ConnectionInfo
		RegisterRelays(...*peer.ConnectionInfo)
		GetConnectionInfo() *peer.ConnectionInfo
		ListenForUpdates() (<-chan UpdateEvent, func())
		// peer certificates
		GetIdentityPublicKey() crypto.PublicKey
		GetPeerCertificate() *object.CertificateResponse
		SetPeerCertificate(*object.CertificateResponse)
		ForgetPeerCertificate()
	}
	localPeer struct {
		keyLock        sync.RWMutex
		primaryPeerKey crypto.PrivateKey
		cids           *ObjectCIDSyncList
		contentTypes   *StringSyncList
		addresses      *StringSyncList
		relays         []*peer.ConnectionInfo
		// peer certificates
		peerCertificateResponse *object.CertificateResponse
		// listeners
		listeners     map[string]chan UpdateEvent
		listenersLock sync.RWMutex
	}
	UpdateEvent string
)

const (
	EventContentTypesUpdated UpdateEvent = "contentTypeUpdated"
	EventCIDsUpdated         UpdateEvent = "cidsUpdated"
	EventAddressesUpdated    UpdateEvent = "addressesUpdated"
	EventRelaysUpdated       UpdateEvent = "relaysUpdated"
	EventIdentityKeyUpdated  UpdateEvent = "identityPublicKeyUpdated"
)

func New() LocalPeer {
	return &localPeer{
		keyLock:       sync.RWMutex{},
		cids:          &ObjectCIDSyncList{},
		contentTypes:  &StringSyncList{},
		addresses:     &StringSyncList{},
		relays:        []*peer.ConnectionInfo{},
		listeners:     map[string]chan UpdateEvent{},
		listenersLock: sync.RWMutex{},
	}
}

func (s *localPeer) SetPeerKey(k crypto.PrivateKey) {
	s.keyLock.Lock()
	s.primaryPeerKey = k
	s.keyLock.Unlock()
}

func (s *localPeer) GetPeerKey() crypto.PrivateKey {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock() //nolint: gocritic
	return s.primaryPeerKey
}

func (s *localPeer) GetIdentityPublicKey() crypto.PublicKey {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock() //nolint: gocritic
	if s.peerCertificateResponse == nil {
		return crypto.EmptyPublicKey
	}
	return s.peerCertificateResponse.Metadata.Signature.Signer
}

func (s *localPeer) GetPeerCertificate() *object.CertificateResponse {
	s.keyLock.Lock()
	defer s.keyLock.Unlock()
	return s.peerCertificateResponse
}

func (s *localPeer) SetPeerCertificate(c *object.CertificateResponse) {
	s.keyLock.Lock()
	s.peerCertificateResponse = c
	s.keyLock.Unlock()
	s.publishUpdate(EventIdentityKeyUpdated)
}

func (s *localPeer) ForgetPeerCertificate() {
	s.keyLock.Lock()
	s.peerCertificateResponse = nil
	s.keyLock.Unlock()
	s.publishUpdate(EventIdentityKeyUpdated)
}

func (s *localPeer) GetAddresses() []string {
	as := s.addresses.List()
	sort.Strings(as)
	return as
}

func (s *localPeer) RegisterAddresses(addresses ...string) {
	for _, h := range addresses {
		s.addresses.Put(h)
	}
	s.publishUpdate(EventAddressesUpdated)
}

func (s *localPeer) GetCIDs() []object.CID {
	return s.cids.List()
}

func (s *localPeer) RegisterCIDs(cids ...object.CID) {
	for _, h := range cids {
		s.cids.Put(h)
	}
	s.publishUpdate(EventCIDsUpdated)
}

func (s *localPeer) GetContentTypes() []string {
	return s.contentTypes.List()
}

func (s *localPeer) RegisterContentTypes(contentTypes ...string) {
	for _, h := range contentTypes {
		s.contentTypes.Put(h)
	}
	s.publishUpdate(EventContentTypesUpdated)
}

func (s *localPeer) GetRelays() []*peer.ConnectionInfo {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock()
	return s.relays
}

func (s *localPeer) RegisterRelays(relays ...*peer.ConnectionInfo) {
	s.keyLock.Lock()
	defer s.keyLock.Unlock()
	s.relays = append(s.relays, relays...)
	s.publishUpdate(EventRelaysUpdated)
}

func (s *localPeer) GetConnectionInfo() *peer.ConnectionInfo {
	return &peer.ConnectionInfo{
		PublicKey: s.GetPeerKey().PublicKey(),
		Addresses: s.GetAddresses(),
		Relays:    s.GetRelays(),
		ObjectFormats: []string{
			"json",
		},
	}
}

func (s *localPeer) publishUpdate(e UpdateEvent) {
	s.listenersLock.RLock()
	defer s.listenersLock.RUnlock()
	for _, l := range s.listeners {
		select {
		case l <- e:
		default:
		}
	}
}

func (s *localPeer) ListenForUpdates() (
	updates <-chan UpdateEvent,
	cancel func(),
) {
	c := make(chan UpdateEvent)
	s.listenersLock.Lock()
	defer s.listenersLock.Unlock()
	id := rand.String(8)
	s.listeners[id] = c
	f := func() {
		s.listenersLock.Lock()
		defer s.listenersLock.Unlock()
		delete(s.listeners, id)
	}
	return c, f
}
