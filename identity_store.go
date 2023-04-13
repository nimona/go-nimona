package nimona

import (
	"fmt"

	"gorm.io/gorm"

	"nimona.io/internal/kv"
)

type IdentityStoreInterface interface {
	NewKeygraph() (*Keygraph, error)
	GetKeygraph(KeygraphID) (*Keygraph, error)
	GetKeyPairs(KeygraphID) (KeyPair, KeyPair, error)
	SignDocument(KeygraphID, *Document) (*Document, error)
}

func NewKeygraphStore(db *gorm.DB) (*KeygraphStore, error) {
	kgStore, err := kv.NewSQLStore[KeygraphID, *Keygraph](db, "keygraphs")
	if err != nil {
		return nil, fmt.Errorf("error creating key graph store: %w", err)
	}
	kpStore, err := kv.NewSQLStore[PublicKey, *KeyPair](db, "keypairs")
	if err != nil {
		return nil, fmt.Errorf("error creating key pair store: %w", err)
	}
	return &KeygraphStore{
		KeygraphStore: kgStore,
		KeyPairStore:  kpStore,
	}, nil
}

type KeygraphStore struct {
	KeygraphStore kv.Store[KeygraphID, *Keygraph]
	KeyPairStore  kv.Store[PublicKey, *KeyPair]
}

func (p *KeygraphStore) NewKeygraph(use string) (*Keygraph, error) {
	kpc, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}

	kpn, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}

	kg := NewKeygraph(kpc.PublicKey, kpn.PublicKey)

	err = p.KeygraphStore.Set(kg.ID(), kg)
	if err != nil {
		return nil, fmt.Errorf("error storing key graph: %w", err)
	}

	err = p.KeyPairStore.Set(kpc.PublicKey, kpc)
	if err != nil {
		return nil, fmt.Errorf("error storing key pair: %w", err)
	}

	err = p.KeyPairStore.Set(kpn.PublicKey, kpn)
	if err != nil {
		return nil, fmt.Errorf("error storing key pair: %w", err)
	}

	return kg, nil
}

func (p *KeygraphStore) GetKeygraph(id KeygraphID) (*Keygraph, error) {
	kg, err := p.KeygraphStore.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error getting key graph: %w", err)
	}

	return kg, nil
}

// nolint: gocritic // unnamed result
func (p *KeygraphStore) GetKeyPairs(id KeygraphID) (*KeyPair, *KeyPair, error) {
	kg, err := p.GetKeygraph(id)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting key graph: %w", err)
	}

	kpc, err := p.KeyPairStore.Get(kg.Keys)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting key pair: %w", err)
	}

	kpn, err := p.KeyPairStore.Get(kg.Next)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting key pair: %w", err)
	}

	return kpc, kpn, nil
}

func (p *KeygraphStore) SignDocument(id KeygraphID, doc *Document) (*Document, error) {
	kpc, _, err := p.GetKeyPairs(id)
	if err != nil {
		return nil, fmt.Errorf("error getting key pairs: %w", err)
	}

	doc = doc.Copy()

	if doc.Metadata.Owner.IsEmpty() {
		doc.Metadata.Owner = id
	}

	doc.Metadata.Signature = NewDocumentSignature(
		kpc.PrivateKey,
		NewDocumentHash(doc),
	)

	return doc, nil
}
