package keychain

import (
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/mockgen -destination=../keychainmock/keychainmock_generated.go -package=keychainmock -source=keychain.go

type (
	Keychain interface {
		GetPrimaryPeerKey() crypto.PrivateKey
		PutPrimaryPeerKey(crypto.PrivateKey)
		GetPrimaryIdentityKey() crypto.PrivateKey
		PutPrimaryIdentityKey(crypto.PrivateKey)
		GetCertificates(crypto.PublicKey) []*peer.Certificate
		PutCertificate(*peer.Certificate)
	}
	memorystore struct {
		keyLock            sync.RWMutex
		certLock           sync.RWMutex
		certs              map[crypto.PublicKey]map[object.Hash]*peer.Certificate
		primaryPeerKey     crypto.PrivateKey
		primaryIdentityKey crypto.PrivateKey
	}
)

func New() Keychain {
	return &memorystore{
		keyLock:  sync.RWMutex{},
		certLock: sync.RWMutex{},
		certs:    map[crypto.PublicKey]map[object.Hash]*peer.Certificate{},
	}
}

func (s *memorystore) PutPrimaryPeerKey(k crypto.PrivateKey) {
	s.keyLock.Lock()
	s.primaryPeerKey = k
	s.keyLock.Unlock()
}

func (s *memorystore) PutPrimaryIdentityKey(k crypto.PrivateKey) {
	s.keyLock.Lock()
	s.primaryIdentityKey = k
	s.keyLock.Unlock()
}

func (s *memorystore) GetPrimaryPeerKey() crypto.PrivateKey {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock() //nolint: gocritic
	return s.primaryPeerKey
}

func (s *memorystore) GetPrimaryIdentityKey() crypto.PrivateKey {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock() //nolint: gocritic
	return s.primaryIdentityKey
}

func (s *memorystore) PutCertificate(c *peer.Certificate) {
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

func (s *memorystore) GetCertificates(
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
