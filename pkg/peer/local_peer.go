package peer

import (
	"sync"
	"time"

	"nimona.io/pkg/bloom"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate $GOBIN/genny -in=$GENERATORS/syncmap/syncmap.go -out=syncmap_string_addresses_generated.go -pkg peer gen "KeyType=string ValueType=Addresses"
//go:generate $GOBIN/genny -in=$GENERATORS/synclist/synclist.go -out=synclist_string_generated.go -pkg peer gen "KeyType=object.Hash"

type (
	Addresses []string
	LocalPeer struct {
		hostname string

		keyLock sync.RWMutex

		peerPrivateKey crypto.PrivateKey
		peerPublicKey  crypto.PublicKey

		identityPrivateKey crypto.PrivateKey
		identityPublicKey  crypto.PublicKey

		certificates []*crypto.Certificate

		addresses     *StringAddressesSyncMap
		contentHashes *ObjectHashSyncList
		contentTypes  []string

		handlerLock             sync.RWMutex
		onAddressesHandlers     []OnAddressesUpdated
		onContentHashesHandlers []OnContentHashesUpdated
	}
	OnAddressesUpdated     func([]string)
	OnContentHashesUpdated func([]object.Hash)
)

func NewLocalPeer(
	hostname string,
	peerPrivateKey crypto.PrivateKey,
) (*LocalPeer, error) {
	return &LocalPeer{
		hostname: hostname,

		peerPrivateKey: peerPrivateKey,
		peerPublicKey:  peerPrivateKey.PublicKey(),

		certificates: []*crypto.Certificate{
			crypto.NewSelfSignedCertificate(peerPrivateKey),
		},

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
func (p *LocalPeer) AddContentHash(hashes ...object.Hash) {
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
func (p *LocalPeer) RemoveContentHash(hashes ...object.Hash) {
	for _, h := range hashes {
		p.contentHashes.Delete(h)
	}
}

func (p *LocalPeer) AddIdentityKey(identityPrivateKey crypto.PrivateKey) error {
	p.keyLock.Lock()
	defer p.keyLock.Unlock()

	p.identityPrivateKey = identityPrivateKey
	p.identityPublicKey = identityPrivateKey.PublicKey()

	p.certificates = append(
		p.certificates,
		crypto.NewCertificate(p.peerPublicKey, identityPrivateKey),
	)

	return nil
}

func (p *LocalPeer) GetPeerPublicKey() crypto.PublicKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	return p.peerPublicKey
}

func (p *LocalPeer) GetPeerPrivateKey() crypto.PrivateKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	return p.peerPrivateKey
}

func (p *LocalPeer) GetIdentityPublicKey() crypto.PublicKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	if p.identityPublicKey.IsEmpty() {
		return p.peerPublicKey
	}
	return p.identityPublicKey
}

func (p *LocalPeer) GetIdentityPrivateKey() crypto.PrivateKey {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	if p.identityPrivateKey == "" {
		return p.peerPrivateKey
	}
	return p.identityPrivateKey
}

func (p *LocalPeer) GetAddresses() []string {
	addrs := []string{}
	p.addresses.Range(func(_ string, addresses *Addresses) bool {
		addrs = append(addrs, []string(*addresses)...)
		return true
	})
	return addrs
}

func (p *LocalPeer) GetContentHashes() []object.Hash {
	hashes := []object.Hash{}
	p.contentHashes.Range(func(hash object.Hash) bool {
		hashes = append(hashes, hash)
		return true
	})
	return hashes
}

func (p *LocalPeer) AddContentTypes(types ...string) {
	p.keyLock.Lock()
	defer p.keyLock.Unlock()
	p.contentTypes = append(p.contentTypes, types...)
}

func (p *LocalPeer) GetContentTypes() []string {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()
	return p.contentTypes
}

func (p *LocalPeer) GetHostname() string {
	return p.hostname
}

// GetSignedPeer returns the local peer info
func (p *LocalPeer) GetSignedPeer() *Peer {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()

	// gather up peer key, certificates, content ids and types
	hs := []string{
		p.peerPublicKey.String(),
	}
	for _, c := range p.GetContentHashes() {
		hs = append(hs, c.String())
	}
	for _, c := range p.contentTypes {
		hs = append(hs, c)
	}
	for _, c := range p.certificates {
		hs = append(hs, c.Signature.Signer.String())
	}

	// TODO cache peer info and reuse
	pi := &Peer{
		Version:      time.Now().UTC().Unix(),
		Bloom:        bloom.New(hs...),
		Addresses:    p.GetAddresses(),
		Certificates: p.certificates,
		ContentTypes: p.contentTypes,
	}

	o := pi.ToObject()
	sig, err := crypto.NewSignature(p.peerPrivateKey, o)
	if err != nil {
		panic(err)
	}

	pi.Signature = sig

	return pi
}

// GetCertificate returns the peer's certificate
func (p *LocalPeer) GetCertificates() []*crypto.Certificate {
	p.keyLock.RLock()
	defer p.keyLock.RUnlock()

	return p.certificates
}
