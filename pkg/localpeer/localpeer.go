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

type (
	LocalPeer interface {
		GetPrimaryPeerKey() crypto.PrivateKey
		PutPrimaryPeerKey(crypto.PrivateKey)
		GetPrimaryIdentityKey() crypto.PrivateKey
		PutPrimaryIdentityKey(crypto.PrivateKey)
		GetCertificates(crypto.PublicKey) []*peer.Certificate
		PutCertificate(*peer.Certificate)
		GetContentHashes() []object.Hash
		PutContentHashes(...object.Hash)
		GetRelays() []*peer.Peer
		PutRelays(...*peer.Peer)
	}
	localPeer struct {
		keyLock            sync.RWMutex
		certLock           sync.RWMutex
		certs              map[crypto.PublicKey]map[object.Hash]*peer.Certificate
		primaryPeerKey     crypto.PrivateKey
		primaryIdentityKey crypto.PrivateKey
		contentHashes      *ObjectHashSyncList
		relays             *PeerPeerSyncList
	}
)

func New() LocalPeer {
	return &localPeer{
		keyLock:       sync.RWMutex{},
		certLock:      sync.RWMutex{},
		certs:         map[crypto.PublicKey]map[object.Hash]*peer.Certificate{},
		contentHashes: &ObjectHashSyncList{},
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
	s.certLock.Lock()
	defer s.certLock.Unlock()
	h := c.ToObject().Hash()
	for _, sub := range c.Metadata.Policy.Subjects {
		if _, ok := s.certs[crypto.PublicKey(sub)]; !ok {
			s.certs[crypto.PublicKey(sub)] = map[object.Hash]*peer.Certificate{}
		}
		s.certs[crypto.PublicKey(sub)][h] = c
	}
}

func (s *localPeer) GetCertificates(
	sub crypto.PublicKey,
) []*peer.Certificate {
	cm, ok := s.certs[sub]
	if !ok {
		return []*peer.Certificate{}
	}
	cs := []*peer.Certificate{}
	for _, c := range cm {
		cs = append(cs, c)
	}
	return cs
}

func (s *localPeer) GetContentHashes() []object.Hash {
	return s.contentHashes.List()
}

func (s *localPeer) PutContentHashes(contentHashes ...object.Hash) {
	for _, h := range contentHashes {
		s.contentHashes.Put(h)
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
