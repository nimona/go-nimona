package localpeer

import (
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/mockgen -destination=../localpeermock/localpeermock_generated.go -package=localpeermock -source=localpeer.go
//go:generate $GOBIN/genny -in=$GENERATORS/synclist/synclist.go -out=contenthashes_generated.go -imp=nimona.io/pkg/object -pkg=localpeer gen "KeyType=object.Hash"
//go:generate $GOBIN/genny -in=$GENERATORS/synclist/synclist.go -out=relays_generated.go -imp=nimona.io/pkg/peer -pkg=localpeer gen "KeyType=*peer.Peer"
//go:generate $GOBIN/genny -in=$GENERATORS/synclist/synclist.go -out=certificates_generated.go -imp=nimona.io/pkg/peer -pkg=localpeer gen "KeyType=*peer.Certificate"
//go:generate $GOBIN/genny -in=$GENERATORS/synclist/synclist.go -out=addresses_generated.go -imp=nimona.io/pkg/peer -pkg=localpeer gen "KeyType=string"

type (
	LocalPeer interface {
		GetPrimaryPeerKey() crypto.PrivateKey
		PutPrimaryPeerKey(crypto.PrivateKey)
		GetPrimaryIdentityKey() crypto.PrivateKey
		PutPrimaryIdentityKey(crypto.PrivateKey)
		GetCertificates() []*peer.Certificate
		PutCertificate(*peer.Certificate)
		GetContentHashes() []object.Hash
		PutContentHashes(...object.Hash)
		GetContentTypes() []string
		PutContentTypes(...string)
		GetAddresses() []string
		PutAddresses(...string)
		GetRelays() []*peer.Peer
		PutRelays(...*peer.Peer)
	}
	localPeer struct {
		keyLock            sync.RWMutex
		primaryPeerKey     crypto.PrivateKey
		primaryIdentityKey crypto.PrivateKey
		contentHashes      *ObjectHashSyncList
		contentTypes       *StringSyncList
		certificates       *PeerCertificateSyncList
		addresses          *StringSyncList
		relays             *PeerPeerSyncList
	}
)

func New() LocalPeer {
	return &localPeer{
		keyLock:       sync.RWMutex{},
		contentHashes: &ObjectHashSyncList{},
		contentTypes:  &StringSyncList{},
		certificates:  &PeerCertificateSyncList{},
		addresses:     &StringSyncList{},
		relays:        &PeerPeerSyncList{},
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

func (s *localPeer) PutCertificate(c *peer.Certificate) {
	s.certificates.Put(c)
}

func (s *localPeer) GetCertificates() []*peer.Certificate {
	return s.certificates.List()
}

func (s *localPeer) GetAddresses() []string {
	return s.addresses.List()
}

func (s *localPeer) PutAddresses(addresses ...string) {
	for _, h := range addresses {
		s.addresses.Put(h)
	}
}

func (s *localPeer) GetContentHashes() []object.Hash {
	return s.contentHashes.List()
}

func (s *localPeer) PutContentHashes(contentHashes ...object.Hash) {
	for _, h := range contentHashes {
		s.contentHashes.Put(h)
	}
}

func (s *localPeer) GetContentTypes() []string {
	return s.contentTypes.List()
}

func (s *localPeer) PutContentTypes(contentTypes ...string) {
	for _, h := range contentTypes {
		s.contentTypes.Put(h)
	}
}

func (s *localPeer) GetRelays() []*peer.Peer {
	return s.relays.List()
}

func (s *localPeer) PutRelays(relays ...*peer.Peer) {
	for _, r := range relays {
		s.relays.Put(r)
	}
}
