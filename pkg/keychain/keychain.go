package keychain

import (
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

var (
	DefaultKeychain = New()
)

//go:generate $GOBIN/mockgen -destination=../keychainmock/keychainmock_generated.go -package=keychainmock -source=keychain.go

type (
	Keychain interface {
		Put(KeyType, crypto.PrivateKey)
		GetPrimaryPeerKey() crypto.PrivateKey
		GetPrimaryIdentityKey() crypto.PrivateKey
		PutCertificate(*peer.Certificate)
		GetCertificates(crypto.PublicKey) []*peer.Certificate
	}
	memorystore struct {
		keyLock            sync.RWMutex
		certLock           sync.RWMutex
		certs              map[crypto.PublicKey]map[object.Hash]*peer.Certificate
		primaryPeerKey     crypto.PrivateKey
		primaryIdentityKey crypto.PrivateKey
	}
)

func Put(kt KeyType, pk crypto.PrivateKey) {
	DefaultKeychain.Put(kt, pk)
}

func GetPrimaryPeerKey() crypto.PrivateKey {
	return DefaultKeychain.GetPrimaryPeerKey()
}

func PutCertificate(cert *peer.Certificate) {
	DefaultKeychain.PutCertificate(cert)
}

func GetCertificates(pk crypto.PublicKey) []*peer.Certificate {
	return DefaultKeychain.GetCertificates(pk)
}

func New() Keychain {
	return &memorystore{
		keyLock:  sync.RWMutex{},
		certLock: sync.RWMutex{},
		certs:    map[crypto.PublicKey]map[object.Hash]*peer.Certificate{},
	}
}

func (s *memorystore) Put(t KeyType, k crypto.PrivateKey) {
	s.keyLock.Lock()
	switch t {
	case PrimaryPeerKey:
		s.primaryPeerKey = k
	case PrimaryIdentityKey:
		s.primaryIdentityKey = k
	}
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
