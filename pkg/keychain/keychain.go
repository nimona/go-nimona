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

type (
	Keychain interface {
		Put(keytype, crypto.PrivateKey)
		List(keytype) []crypto.PrivateKey
		ListPublicKeys(keytype) []crypto.PublicKey
		GetPrimaryPeerKey() crypto.PrivateKey
		PutCertificate(*peer.Certificate)
		GetCertificates(crypto.PublicKey) []*peer.Certificate
	}
	memorystore struct {
		keyLock        sync.RWMutex
		keys           map[keytype]map[crypto.PrivateKey]struct{}
		certLock       sync.RWMutex
		certs          map[crypto.PublicKey]map[object.Hash]*peer.Certificate
		primaryPeerKey crypto.PrivateKey
	}
)

func Put(kt keytype, pk crypto.PrivateKey) {
	DefaultKeychain.Put(kt, pk)
}

func List(kt keytype) []crypto.PrivateKey {
	return DefaultKeychain.List(kt)
}

func ListPublicKeys(kt keytype) []crypto.PublicKey {
	return DefaultKeychain.ListPublicKeys(kt)
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
		keyLock: sync.RWMutex{},
		keys: map[keytype]map[crypto.PrivateKey]struct{}{
			PeerKey:     {},
			IdentityKey: {},
		},
		certLock: sync.RWMutex{},
		certs:    map[crypto.PublicKey]map[object.Hash]*peer.Certificate{},
	}
}

func (s *memorystore) Put(t keytype, k crypto.PrivateKey) {
	s.keyLock.Lock()
	switch t {
	case PrimaryPeerKey:
		s.primaryPeerKey = k
		s.keys[PeerKey][k] = struct{}{}
	default:
		s.keys[t][k] = struct{}{}
	}
	s.keyLock.Unlock()
}

func (s *memorystore) List(t keytype) []crypto.PrivateKey {
	s.keyLock.RLock()
	ks := []crypto.PrivateKey{}
	for k := range s.keys[t] {
		ks = append(ks, k)
	}
	s.keyLock.RUnlock()
	return ks
}

func (s *memorystore) ListPublicKeys(t keytype) []crypto.PublicKey {
	s.keyLock.RLock()
	pks := []crypto.PublicKey{}
	for k := range s.keys[t] {
		pks = append(pks, k.PublicKey())
	}
	s.keyLock.RUnlock()
	return pks
}

func (s *memorystore) GetPrimaryPeerKey() crypto.PrivateKey {
	s.keyLock.RLock()
	defer s.keyLock.RUnlock() //nolint: gocritic
	return s.primaryPeerKey
}

func (s *memorystore) PutCertificate(c *peer.Certificate) {
	s.certLock.Lock()
	defer s.certLock.Unlock()
	h := c.ToObject().Hash()
	for _, sub := range c.Policy.Subjects {
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
