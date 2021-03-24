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
		GetPrimaryPeerKey() crypto.PrivateKey
		PutPrimaryPeerKey(crypto.PrivateKey)
		GetPrimaryIdentityKey() crypto.PrivateKey
		PutPrimaryIdentityKey(crypto.PrivateKey)
		GetCertificates() []*object.Certificate
		PutCertificate(*object.Certificate)
		GetCIDs() []object.CID
		PutCIDs(...object.CID)
		GetContentTypes() []string
		PutContentTypes(...string)
		GetAddresses() []string
		PutAddresses(...string)
		GetRelays() []*peer.ConnectionInfo
		PutRelays(...*peer.ConnectionInfo)
		ConnectionInfo() *peer.ConnectionInfo
		ListenForUpdates() (<-chan UpdateEvent, func())
	}
	localPeer struct {
		keyLock            sync.RWMutex
		primaryPeerKey     crypto.PrivateKey
		primaryIdentityKey crypto.PrivateKey
		cids               *ObjectCIDSyncList
		contentTypes       *StringSyncList
		certificates       *ObjectCertificateSyncList
		addresses          *StringSyncList
		relays             []*peer.ConnectionInfo
		listeners          map[string]chan UpdateEvent
		listenersLock      sync.RWMutex
	}
	UpdateEvent string
)

const (
	EventContentTypesUpdated       UpdateEvent = "contentTypeUpdated"
	EventCIDsUpdated               UpdateEvent = "cidsUpdated"
	EventAddressesUpdated          UpdateEvent = "addressesUpdated"
	EventRelaysUpdated             UpdateEvent = "relaysUpdated"
	EventPrimaryIdentityKeyUpdated UpdateEvent = "primaryIdentityKeyUpdated"
)

func New() LocalPeer {
	return &localPeer{
		keyLock:       sync.RWMutex{},
		cids:          &ObjectCIDSyncList{},
		contentTypes:  &StringSyncList{},
		certificates:  &ObjectCertificateSyncList{},
		addresses:     &StringSyncList{},
		relays:        []*peer.ConnectionInfo{},
		listeners:     map[string]chan UpdateEvent{},
		listenersLock: sync.RWMutex{},
	}
}

func (s *localPeer) PutPrimaryPeerKey(k crypto.PrivateKey) {
	s.keyLock.Lock()
	s.primaryPeerKey = k
	s.keyLock.Unlock()
}

func (s *localPeer) PutPrimaryIdentityKey(k crypto.PrivateKey) {
	s.keyLock.Lock()
	s.primaryIdentityKey = k
	s.keyLock.Unlock()
	s.publishUpdate(EventPrimaryIdentityKeyUpdated)
}

func (s *localPeer) GetPrimaryPeerKey() crypto.PrivateKey {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock() //nolint: gocritic
	return s.primaryPeerKey
}

func (s *localPeer) GetPrimaryIdentityKey() crypto.PrivateKey {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock() //nolint: gocritic
	return s.primaryIdentityKey
}

func (s *localPeer) PutCertificate(c *object.Certificate) {
	s.certificates.Put(c)
}

func (s *localPeer) GetCertificates() []*object.Certificate {
	return s.certificates.List()
}

func (s *localPeer) GetAddresses() []string {
	as := s.addresses.List()
	sort.Strings(as)
	return as
}

func (s *localPeer) PutAddresses(addresses ...string) {
	for _, h := range addresses {
		s.addresses.Put(h)
	}
	s.publishUpdate(EventAddressesUpdated)
}

func (s *localPeer) GetCIDs() []object.CID {
	return s.cids.List()
}

func (s *localPeer) PutCIDs(cids ...object.CID) {
	for _, h := range cids {
		s.cids.Put(h)
	}
	s.publishUpdate(EventCIDsUpdated)
}

func (s *localPeer) GetContentTypes() []string {
	return s.contentTypes.List()
}

func (s *localPeer) PutContentTypes(contentTypes ...string) {
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

func (s *localPeer) PutRelays(relays ...*peer.ConnectionInfo) {
	s.keyLock.Lock()
	defer s.keyLock.Unlock()
	s.relays = append(s.relays, relays...)
	s.publishUpdate(EventRelaysUpdated)
}

func (s *localPeer) ConnectionInfo() *peer.ConnectionInfo {
	return &peer.ConnectionInfo{
		PublicKey: s.GetPrimaryPeerKey().PublicKey(),
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
