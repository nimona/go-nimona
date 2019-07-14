package peer

import (
	"crypto/ecdsa"
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

//go:generate $GOBIN/genny -in=../../internal/generator/syncmap/syncmap.go -out=syncmap_addresses_generated.go -pkg peer gen "KeyType=string ValueType=Addresses"
//go:generate $GOBIN/genny -in=../../internal/generator/syncmap/syncmap.go -out=syncmap_addresses_handlers_generated.go -pkg peer gen "KeyType=string ValueType=bool"

//go:generate $GOBIN/genny -in=../../internal/generator/syncmap/syncmap.go -out=syncmap_on_addresses_generated.go -pkg peer gen "KeyType=OnAddressesUpdated ValueType=bool"
//go:generate $GOBIN/genny -in=../../internal/generator/syncmap/syncmap.go -out=syncmap_on_content_hashes_generated.go -pkg peer gen "KeyType=OnContentHashesUpdated ValueType=bool"

type (
	Addresses []string
	Peer      struct {
		fingerprint crypto.Fingerprint
		hostname    string

		keyLock     sync.RWMutex
		key         *crypto.PrivateKey
		identityKey *crypto.PrivateKey

		addresses     *StringAddressesSyncMap
		contentHashes *StringBoolSyncMap

		onAddressesHandlers     *OnAddressesUpdatedBoolSyncMap
		onContentHashesHandlers *OnContentHashesUpdatedBoolSyncMap
	}
	OnAddressesUpdated     func([]string)
	OnContentHashesUpdated func([]string)
)

func NewPeer(hostname string, key *crypto.PrivateKey) (
	*Peer, error) {
	if key == nil {
		return nil, ErrMissingKey
	}

	if _, ok := key.Key().(*ecdsa.PrivateKey); !ok {
		return nil, ErrECDSAPrivateKeyRequired
	}

	return &Peer{
		fingerprint: key.Fingerprint(),
		hostname:    hostname,
		key:         key,

		addresses:     &StringAddressesSyncMap{},
		contentHashes: &StringBoolSyncMap{},

		onAddressesHandlers:     &OnAddressesUpdatedBoolSyncMap{},
		onContentHashesHandlers: &OnContentHashesUpdatedBoolSyncMap{},
	}, nil
}

func (l *Peer) OnAddressesUpdated(h OnAddressesUpdated) {
	t := true
	l.onAddressesHandlers.Put(h, &t)
}

func (l *Peer) OnContentHashesUpdated(h OnContentHashesUpdated) {
	t := true
	l.onContentHashesHandlers.Put(h, &t)
}

func (l *Peer) AddAddress(protocol string, addrs []string) {
	a := Addresses(addrs)
	l.addresses.Put(protocol, &a)
	all := l.GetAddresses()
	l.onAddressesHandlers.Range(func(h OnAddressesUpdated, _ *bool) bool {
		h(all)
		return true
	})
}

// AddContentHash that should be published with the peer info
func (l *Peer) AddContentHash(hashes ...string) {
	for _, h := range hashes {
		t := true
		l.contentHashes.Put(h, &t)
	}
	all := l.GetContentHashes()
	l.onContentHashesHandlers.Range(func(h OnContentHashesUpdated, _ *bool) bool {
		h(all)
		return true
	})
}

// RemoveContentHash from the peer info
func (l *Peer) RemoveContentHash(hashes ...string) {
	for _, h := range hashes {
		t := false
		l.contentHashes.Put(h, &t)
	}
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

func (l *Peer) GetAddresses() []string {
	addrs := []string{}
	l.addresses.Range(func(_ string, addresses *Addresses) bool {
		addrs = append(addrs, []string(*addresses)...)
		return true
	})
	return addrs
}

func (l *Peer) GetContentHashes() []string {
	hashes := []string{}
	l.contentHashes.Range(func(hash string, ok *bool) bool {
		if *ok == true {
			hashes = append(hashes, hash)
		}
		return true
	})
	return hashes
}

// GetPeerInfo returns the local peer info
func (l *Peer) GetPeerInfo() *peer.PeerInfo {
	// TODO cache peer info and reuse
	p := &peer.PeerInfo{
		Addresses:  l.GetAddresses(),
		ContentIDs: l.GetContentHashes(),
	}

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
