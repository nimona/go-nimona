package localpeer

import (
	"sync"

	"nimona.io/internal/rand"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate mockgen -destination=../localpeermock/localpeermock_generated.go -package=localpeermock -source=localpeer.go
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=cids_generated.go -imp=nimona.io/pkg/object -pkg=localpeer gen "KeyType=object.CID"
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=certificates_generated.go -imp=nimona.io/pkg/peer -pkg=localpeer gen "KeyType=*object.Certificate"

type (
	LocalPeer interface {
		// TODO(geoah) move to object store
		GetCIDs() []object.CID
		RegisterCIDs(...object.CID)
		// local peer information
		GetPeerKey() crypto.PrivateKey
		SetPeerKey(crypto.PrivateKey)
		GetIdentityPublicKey() crypto.PublicKey
		GetPeerCertificate() *object.CertificateResponse
		SetPeerCertificate(*object.CertificateResponse)
		ForgetPeerCertificate()
		ListenForUpdates() (<-chan UpdateEvent, func())
	}
	localPeer struct {
		keyLock                 sync.RWMutex
		primaryPeerKey          crypto.PrivateKey
		cids                    *ObjectCIDSyncList
		peerCertificateResponse *object.CertificateResponse
		listeners               map[string]chan UpdateEvent
		listenersLock           sync.RWMutex
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

func (s *localPeer) GetCIDs() []object.CID {
	return s.cids.List()
}

func (s *localPeer) RegisterCIDs(cids ...object.CID) {
	for _, h := range cids {
		s.cids.Put(h)
	}
	s.publishUpdate(EventCIDsUpdated)
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
