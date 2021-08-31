package keystream

import (
	"fmt"
	"sync"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/did"
	"nimona.io/pkg/keystore"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/tilde"
)

type (
	// Controller deals with the key management and event transitions for a
	// single key stream
	Controller interface {
		Rotate() (*Rotation, error)
		Delegate(DelegateSeal) (*DelegationInteraction, error)
		CurrentKey() crypto.PrivateKey
		// TODO should this be returning a pointer or copy?
		GetKeyStream() *State
	}
	controller struct {
		mutex             sync.RWMutex
		keyStore          keystore.KeyStore
		objectStore       objectstore.Store
		state             *State
		currentPrivateKey crypto.PrivateKey
		newKey            func() (crypto.PrivateKey, error)
	}
)

func RestoreController(
	keyStreamRootHash tilde.Digest,
	keyStore keystore.KeyStore,
	objectStore objectstore.Store,
) (*controller, error) {
	eventStream, err := objectStore.GetByStream(
		keyStreamRootHash,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get keystream objects, %w", err)
	}

	keyStream, err := FromStream(eventStream)
	if err != nil {
		return nil, fmt.Errorf("unable to create state, %w", err)
	}

	pk, err := keyStore.GetKey(keyStream.ActiveKey.Hash())
	if err != nil {
		return nil, fmt.Errorf("unable to get private active key, %w", err)
	}

	c := &controller{
		mutex:             sync.RWMutex{},
		keyStore:          keyStore,
		objectStore:       objectStore,
		state:             keyStream,
		currentPrivateKey: *pk,
		newKey:            crypto.NewEd25519PrivateKey,
	}
	return c, nil
}

func NewController(
	owner did.DID,
	keyStore keystore.KeyStore,
	objectStore objectstore.Store,
	delegatorSeal *DelegatorSeal,
) (*controller, error) {
	k0, err := crypto.NewEd25519PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to generate key, %w", err)
	}
	k1, err := crypto.NewEd25519PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to generate key, %w", err)
	}
	err = keyStore.PutKey(k0)
	if err != nil {
		return nil, fmt.Errorf("unable to put key, %w", err)
	}
	err = keyStore.PutKey(k1)
	if err != nil {
		return nil, fmt.Errorf("unable to put key, %w", err)
	}
	inceptionEvent := &Inception{
		Metadata: object.Metadata{
			Owner:    owner,
			Sequence: 0,
		},
		Version:       Version,
		Key:           k0.PublicKey(),
		NextKeyDigest: k1.PublicKey().Hash(),
		DelegatorSeal: delegatorSeal,
	}
	inceptionObject, err := object.Marshal(inceptionEvent)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal object, %w", err)
	}
	err = object.Sign(k0, inceptionObject)
	if err != nil {
		return nil, fmt.Errorf("unable to sign object, %w", err)
	}
	err = objectStore.Put(inceptionObject)
	if err != nil {
		return nil, fmt.Errorf("unable to put object, %w", err)
	}
	keyStream, err := FromStream(
		object.NewReadCloserFromObjects(
			[]*object.Object{
				inceptionObject,
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new state, %w", err)
	}

	c := &controller{
		mutex:             sync.RWMutex{},
		keyStore:          keyStore,
		objectStore:       objectStore,
		state:             keyStream,
		currentPrivateKey: k0,
		newKey:            crypto.NewEd25519PrivateKey,
	}

	return c, nil
}

func (c *controller) CurrentKey() crypto.PrivateKey {
	return c.currentPrivateKey
}

func (c *controller) GetKeyStream() *State {
	return c.state
}

func (c *controller) Rotate() (*Rotation, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	newNextKey, err := crypto.NewEd25519PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to create a new key, %w", err)
	}

	newCurrentKey, err := c.keyStore.GetKey(c.state.NextKeyDigest)
	if err != nil {
		return nil, fmt.Errorf("unable to get next private key, %w", err)
	}

	err = c.keyStore.PutKey(newNextKey)
	if err != nil {
		return nil, fmt.Errorf("unable to store new next private key, %w", err)
	}

	r := &Rotation{
		Metadata: object.Metadata{
			Sequence: c.state.Sequence + 1,
		},
		Version:       Version,
		Key:           newCurrentKey.PublicKey(),
		NextKeyDigest: newNextKey.PublicKey().Hash(),
	}

	c.currentPrivateKey = *newCurrentKey

	err = r.apply(c.state)
	if err != nil {
		return nil, fmt.Errorf("unable to apply rotation on state, %w", err)
	}

	return r, nil
}

func (c *controller) Delegate(
	ds DelegateSeal,
) (*DelegationInteraction, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	d := &DelegationInteraction{
		Metadata: object.Metadata{
			Sequence: c.state.Sequence + 1,
		},
		Version:      Version,
		DelegateSeal: ds,
	}

	err := d.apply(c.state)
	if err != nil {
		return nil, fmt.Errorf("unable to apply delegation on state, %w", err)
	}

	return d, nil
}
