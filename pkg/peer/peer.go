package peer

import (
	"crypto/ecdsa"
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

type Peer struct {
	fingerprint crypto.Fingerprint
	hostname    string

	keyLock     sync.RWMutex
	key         *crypto.PrivateKey
	identityKey *crypto.PrivateKey

	addressesLock sync.RWMutex
	addresses     []string

	// TODO replace with generated string-bool syncmap
	contentHashesLock sync.RWMutex
	contentHashes     map[string]bool // map[hash]publishable
}

func NewPeer(hostname string, key *crypto.PrivateKey) (
	*Peer, error) {
	if key == nil {
		return nil, ErrMissingKey
	}

	if _, ok := key.Key().(*ecdsa.PrivateKey); !ok {
		return nil, ErrECDSAPrivateKeyRequired
	}

	return &Peer{
		fingerprint:   key.Fingerprint(),
		hostname:      hostname,
		key:           key,
		addresses:     []string{},
		contentHashes: map[string]bool{},
	}, nil
}

func (l *Peer) AddAddress(addrs ...string) {
	l.addressesLock.Lock()
	if l.addresses == nil {
		l.addresses = []string{}
	}
	l.addresses = append(l.addresses, addrs...)
	l.addressesLock.Unlock()
}

// AddContentHash that should be published with the peer info
func (l *Peer) AddContentHash(hashes ...string) {
	l.contentHashesLock.Lock()
	for _, hash := range hashes {
		l.contentHashes[hash] = true
	}
	l.contentHashesLock.Unlock()
}

// RemoveContentHash from the peer info
func (l *Peer) RemoveContentHash(hashes ...string) {
	l.contentHashesLock.Lock()
	for _, hash := range hashes {
		delete(l.contentHashes, hash)
	}
	l.contentHashesLock.Unlock()
}

func (l *Peer) AddIdentityKey(identityKey *crypto.PrivateKey) error {
	l.keyLock.Lock()
	defer l.keyLock.Unlock()

	pko := l.key.PublicKey.ToObject()
	sig, err := crypto.NewSignature(
		identityKey,
		crypto.AlgorithmObjectHash,
		pko,
	)
	if err != nil {
		return err
	}

	l.identityKey = identityKey
	l.key.PublicKey.Signature = sig

	return nil
}

func (l *Peer) GetPeerKey() *crypto.PrivateKey {
	l.keyLock.RLock()
	defer l.keyLock.RUnlock()

	return l.key
}

// GetPeerInfo returns the local peer info
func (l *Peer) GetPeerInfo() *peer.PeerInfo {
	// TODO cache peer info and reuse
	p := &peer.PeerInfo{}

	l.addressesLock.RLock()
	defer l.contentHashesLock.RUnlock()

	// TODO Check all the transports for addresses
	addresses := make([]string, len(l.addresses))
	for i, a := range l.addresses {
		addresses[i] = a
	}
	p.Addresses = addresses
	l.addressesLock.RUnlock()

	l.contentHashesLock.RLock()
	hashes := []string{}
	for hash, publishable := range l.contentHashes {
		if !publishable {
			continue
		}
		hashes = append(hashes, hash)
	}
	p.ContentIDs = hashes

	o := p.ToObject()
	if err := crypto.Sign(o, l.GetPeerKey()); err != nil {
		panic(err)
	}
	if err := p.FromObject(o); err != nil {
		panic(err)
	}
	return p
}

func (l *Peer) GetFingerprint() crypto.Fingerprint {
	return l.fingerprint
}

func (l *Peer) GetHostname() string {
	return l.hostname
}
