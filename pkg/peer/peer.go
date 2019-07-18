package peer

import (
	"crypto/ecdsa"
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

//go:generate $GOBIN/genny -in=../../internal/generator/syncmap/syncmap.go -out=syncmap_string_addresses_generated.go -pkg peer gen "KeyType=string ValueType=Addresses"
//go:generate $GOBIN/genny -in=../../internal/generator/synclist/synclist.go -out=synclist_string_generated.go -pkg peer gen "KeyType=string"

//go:generate $GOBIN/genny -in=../../internal/generator/synclist/synclist.go -out=synclist_on_addresses_generated.go -pkg peer gen "KeyType=OnAddressesUpdated"
//go:generate $GOBIN/genny -in=../../internal/generator/synclist/synclist.go -out=synclist_on_content_hashes_generated.go -pkg peer gen "KeyType=OnContentHashesUpdated"

type (
	Addresses []string
	Peer      struct {
		fingerprint crypto.Fingerprint
		hostname    string

		keyLock     sync.RWMutex
		key         *crypto.PrivateKey
		identityKey *crypto.PrivateKey

		addresses     *StringAddressesSyncMap
		contentHashes *StringSyncList

		onAddressesHandlers     *OnAddressesUpdatedSyncList
		onContentHashesHandlers *OnContentHashesUpdatedSyncList
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
		contentHashes: &StringSyncList{},

		onAddressesHandlers:     &OnAddressesUpdatedSyncList{},
		onContentHashesHandlers: &OnContentHashesUpdatedSyncList{},
	}, nil
}

func (p *Peer) OnAddressesUpdated(h OnAddressesUpdated) {
	p.onAddressesHandlers.Put(h)
}

func (p *Peer) OnContentHashesUpdated(h OnContentHashesUpdated) {
	p.onContentHashesHandlers.Put(h)
}

func (p *Peer) AddAddress(protocol string, addrs []string) {
	a := Addresses(addrs)
	p.addresses.Put(protocol, &a)
	all := p.GetAddresses()
	p.onAddressesHandlers.Range(func(h OnAddressesUpdated) bool {
		h(all)
		return true
	})
}

// AddContentHash that should be published with the peer info
func (p *Peer) AddContentHash(hashes ...string) {
	for _, h := range hashes {
		p.contentHashes.Put(h)
	}
	all := p.GetContentHashes()
	p.onContentHashesHandlers.Range(func(h OnContentHashesUpdated) bool {
		h(all)
		return true
	})
}

// RemoveContentHash from the peer info
func (p *Peer) RemoveContentHash(hashes ...string) {
	for _, h := range hashes {
		p.contentHashes.Delete(h)
	}
}

func (p *Peer) AddIdentityKey(identityKey *crypto.PrivateKey) error {
	p.keyLock.Lock()
	defer p.keyLock.Unlock()

	pko := p.key.PublicKey.ToObject()
	sig, err := crypto.NewSignature(
		identityKey,
		crypto.AlgorithmObjectHash,
		pko,
	)
	if err != nil {
		return err
	}

	p.identityKey = identityKey
	p.key.PublicKey.Signature = sig

	return nil
}

func (p *Peer) GetPeerKey() *crypto.PrivateKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	return p.key
}

func (p *Peer) GetAddresses() []string {
	addrs := []string{}
	p.addresses.Range(func(_ string, addresses *Addresses) bool {
		addrs = append(addrs, []string(*addresses)...)
		return true
	})
	return addrs
}

func (p *Peer) GetContentHashes() []string {
	hashes := []string{}
	p.contentHashes.Range(func(hash string) bool {
		hashes = append(hashes, hash)
		return true
	})
	return hashes
}

// GetPeerInfo returns the local peer info
func (p *Peer) GetPeerInfo() *peer.PeerInfo {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()

	// TODO cache peer info and reuse
	pi := &peer.PeerInfo{
		Addresses:  p.GetAddresses(),
		ContentIDs: p.GetContentHashes(),
	}

	o := pi.ToObject()
	if err := crypto.Sign(o, p.GetPeerKey()); err != nil {
		panic(err)
	}
	if err := pi.FromObject(o); err != nil {
		panic(err)
	}
	return pi
}

func (p *Peer) GetFingerprint() crypto.Fingerprint {
	return p.fingerprint
}

func (p *Peer) GetHostname() string {
	return p.hostname
}
