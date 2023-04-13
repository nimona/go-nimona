package nimona

import (
	"fmt"

	"gorm.io/gorm"

	"nimona.io/internal/kv"
)

type IdentityStoreInterface interface {
	NewKeyGraph() (*KeyGraph, error)
	GetKeyGraph(KeyGraphID) (*KeyGraph, error)
	GetKeyPairs(KeyGraphID) (KeyPair, KeyPair, error)
	SignDocument(KeyGraphID, *Document) (*Document, error)
}

func NewKeyGraphStore(db *gorm.DB) (*KeyGraphStore, error) {
	kgStore, err := kv.NewSQLStore[KeyGraphID, *KeyGraph](db, "keygraphs")
	if err != nil {
		return nil, fmt.Errorf("error creating key graph store: %w", err)
	}
	kpStore, err := kv.NewSQLStore[PublicKey, *KeyPair](db, "keypairs")
	if err != nil {
		return nil, fmt.Errorf("error creating key pair store: %w", err)
	}
	return &KeyGraphStore{
		KeyGraphStore: kgStore,
		KeyPairStore:  kpStore,
	}, nil
}

type KeyGraphStore struct {
	KeyGraphStore kv.Store[KeyGraphID, *KeyGraph]
	KeyPairStore  kv.Store[PublicKey, *KeyPair]
}

func (p *KeyGraphStore) NewKeyGraph(use string) (*KeyGraph, error) {
	kpc, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}

	kpn, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}

	kg := NewKeyGraph(kpc.PublicKey, kpn.PublicKey)

	err = p.KeyGraphStore.Set(kg.ID(), kg)
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

func (p *KeyGraphStore) GetKeyGraph(id KeyGraphID) (*KeyGraph, error) {
	kg, err := p.KeyGraphStore.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error getting key graph: %w", err)
	}

	return kg, nil
}

// nolint: gocritic // unnamed result
func (p *KeyGraphStore) GetKeyPairs(id KeyGraphID) (*KeyPair, *KeyPair, error) {
	kg, err := p.GetKeyGraph(id)
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

func (p *KeyGraphStore) SignDocument(id KeyGraphID, doc *Document) (*Document, error) {
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
