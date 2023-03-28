package nimona

import (
	"fmt"

	"gorm.io/gorm"

	"nimona.io/internal/kv"
)

type IdentityStoreInterface interface {
	NewIdentity() (*Identity, error)
	GetKeyGraph(*Identity) (*KeyGraph, error)
	GetKeyPairs(*Identity) (KeyPair, KeyPair, error)
	SignDocument(*Identity, *Document) (*Document, error)
}

func NewIdentityStore(db *gorm.DB) (*IdentityStore, error) {
	kgStore, err := kv.NewSQLStore[Identity, KeyGraph](db)
	if err != nil {
		return nil, fmt.Errorf("error creating key graph store: %w", err)
	}
	kpStore, err := kv.NewSQLStore[PublicKey, KeyPair](db)
	if err != nil {
		return nil, fmt.Errorf("error creating key pair store: %w", err)
	}
	return &IdentityStore{
		IdentityStore: kgStore,
		KeyPairStore:  kpStore,
	}, nil
}

type IdentityStore struct {
	IdentityStore kv.Store[Identity, KeyGraph]
	KeyPairStore  kv.Store[PublicKey, KeyPair]
}

func (p *IdentityStore) NewIdentity(use string) (*Identity, error) {
	kpc, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}

	kpn, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}

	kg := NewKeyGraph(kpc.PublicKey, kpn.PublicKey)
	id := NewIdentity(use, kg)

	err = p.IdentityStore.Set(*id, kg)
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

	return id, nil
}

func (p *IdentityStore) GetKeyGraph(id *Identity) (*KeyGraph, error) {
	kg, err := p.IdentityStore.Get(*id)
	if err != nil {
		return nil, fmt.Errorf("error getting key graph: %w", err)
	}

	return kg, nil
}

// nolint: gocritic // unnamed result
func (p *IdentityStore) GetKeyPairs(id *Identity) (*KeyPair, *KeyPair, error) {
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
