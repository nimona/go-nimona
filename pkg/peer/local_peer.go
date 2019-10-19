package peer

import (
	"crypto/ecdsa"
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate $GOBIN/genny -in=$GENERATORS/syncmap/syncmap.go -out=syncmap_string_addresses_generated.go -pkg peer gen "KeyType=string ValueType=Addresses"
//go:generate $GOBIN/genny -in=$GENERATORS/synclist/synclist.go -out=synclist_string_generated.go -pkg peer gen "KeyType=*object.Hash"

type (
	Addresses []string
	LocalPeer struct {
		fingerprint crypto.Fingerprint
		hostname    string

		keyLock     sync.RWMutex
		key         *crypto.PrivateKey
		identityKey *crypto.PrivateKey

		addresses     *StringAddressesSyncMap
		contentHashes *ObjectHashSyncList

		handlerLock             sync.RWMutex
		onAddressesHandlers     []OnAddressesUpdated
		onContentHashesHandlers []OnContentHashesUpdated
	}
	OnAddressesUpdated     func([]string)
	OnContentHashesUpdated func([]*object.Hash)
)

func NewLocalPeer(
	hostname string,
	key *crypto.PrivateKey,
) (*LocalPeer, error) {
	if key == nil {
		return nil, ErrMissingKey
	}

	if _, ok := key.Key().(*ecdsa.PrivateKey); !ok {
		return nil, ErrECDSAPrivateKeyRequired
	}

	return &LocalPeer{
		fingerprint: key.Fingerprint(),
		hostname:    hostname,
		key:         key,

		addresses:     &StringAddressesSyncMap{},
		contentHashes: &ObjectHashSyncList{},

		onAddressesHandlers:     []OnAddressesUpdated{},
		onContentHashesHandlers: []OnContentHashesUpdated{},
	}, nil
}

func (p *LocalPeer) OnAddressesUpdated(h OnAddressesUpdated) {
	p.handlerLock.Lock()
	p.onAddressesHandlers = append(p.onAddressesHandlers, h)
	p.handlerLock.Unlock()
}

func (p *LocalPeer) OnContentHashesUpdated(h OnContentHashesUpdated) {
	p.handlerLock.Lock()
	p.onContentHashesHandlers = append(p.onContentHashesHandlers, h)
	p.handlerLock.Unlock()
}

func (p *LocalPeer) AddAddress(protocol string, addrs []string) {
	a := Addresses(addrs)
	p.addresses.Put(protocol, &a)
	all := p.GetAddresses()
	p.handlerLock.RLock()
	for _, h := range p.onAddressesHandlers {
		go h(all)
	}
	p.handlerLock.RUnlock()
}

// AddContentHash that should be published with the peer info
func (p *LocalPeer) AddContentHash(hashes ...*object.Hash) {
	for _, h := range hashes {
		p.contentHashes.Put(h)
	}
	all := p.GetContentHashes()
	p.handlerLock.RLock()
	for _, h := range p.onContentHashesHandlers {
		go h(all)
	}
	p.handlerLock.RUnlock()
}

// RemoveContentHash from the peer info
func (p *LocalPeer) RemoveContentHash(hashes ...*object.Hash) {
	for _, h := range hashes {
		p.contentHashes.Delete(h)
	}
}

func (p *LocalPeer) AddIdentityKey(identityKey *crypto.PrivateKey) error {
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

func (p *LocalPeer) GetPeerKey() *crypto.PublicKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	return p.key.PublicKey
}

func (p *LocalPeer) GetPeerPrivateKey() *crypto.PrivateKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	return p.key
}

func (p *LocalPeer) GetIdentityKey() *crypto.PublicKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	if p.identityKey == nil {
		return p.key.PublicKey
	}
	return p.identityKey.PublicKey
}

func (p *LocalPeer) GetIdentityPrivateKey() *crypto.PrivateKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	if p.identityKey == nil {
		return p.key
	}
	return p.identityKey
}

func (p *LocalPeer) GetAddress() string {
	return "peer:" + p.key.Fingerprint().String()
}

func (p *LocalPeer) GetAddresses() []string {
	addrs := []string{}
	p.addresses.Range(func(_ string, addresses *Addresses) bool {
		addrs = append(addrs, []string(*addresses)...)
		return true
	})
	return addrs
}

func (p *LocalPeer) GetContentHashes() []*object.Hash {
	hashes := []*object.Hash{}
	p.contentHashes.Range(func(hash *object.Hash) bool {
		hashes = append(hashes, hash)
		return true
	})
	return hashes
}

func (p *LocalPeer) GetFingerprint() crypto.Fingerprint {
	return p.fingerprint
}

func (p *LocalPeer) GetHostname() string {
	return p.hostname
}

// GetSignedPeer returns the local peer info
func (p *LocalPeer) GetSignedPeer() *Peer {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()

	// TODO cache peer info and reuse
	pi := &Peer{
		Addresses: p.GetAddresses(),
	}

	o := pi.ToObject()
	if err := crypto.Sign(o, p.GetPeerPrivateKey()); err != nil {
		panic(err)
	}
	if err := pi.FromObject(o); err != nil {
		panic(err)
	}
	return pi
}
