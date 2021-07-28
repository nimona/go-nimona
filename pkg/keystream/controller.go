package keystream

import (
	"fmt"
	"sync"

	"github.com/xujiajun/nutsdb"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

const (
	// the key for each private key in this bucket is the public key's hash
	privateKeysBucket = "keystream_private_keys"
)

type (
	// Controller deals with the key management and event transitions for our
	// own key stream
	Controller struct {
		mutex     sync.RWMutex
		kvStore   *nutsdb.DB
		state     *KeyStream
		activeKey crypto.PrivateKey
		newKey    func() (crypto.PrivateKey, error)
	}
)

func NewController(
	kvStore *nutsdb.DB,
	eventStream object.ReadCloser,
) (*Controller, error) {
	s, err := FromStream(eventStream)
	if err != nil {
		return nil, fmt.Errorf("unable to create state, %w", err)
	}

	// TODO deal with first time initialization
	// TODO check that we have the next key as well

	pk, err := getPrivateKey(getPublicKeyHash(s.ActiveKey), kvStore)
	if err != nil {
		return nil, fmt.Errorf("unable to get private active key, %w", err)
	}

	c := &Controller{
		mutex:     sync.RWMutex{},
		kvStore:   kvStore,
		state:     s,
		activeKey: *pk,
		newKey:    crypto.NewEd25519PrivateKey,
	}

	return c, nil
}

func (c *Controller) Rotate() (*Rotation, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	newNextKey, err := crypto.NewEd25519PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to create a new key, %w", err)
	}

	newCurrentKey, err := getPrivateKey(c.state.NextKeyHash, c.kvStore)
	if err != nil {
		return nil, fmt.Errorf("unable to get next private key, %w", err)
	}

	err = storePrivateKey(newNextKey, c.kvStore)
	if err != nil {
		return nil, fmt.Errorf("unable to store new next private key, %w", err)
	}

	r := &Rotation{
		Metadata:      object.Metadata{},
		Version:       Version,
		Key:           newCurrentKey.PublicKey(),
		NextKeyDigest: getPublicKeyHash(newNextKey.PublicKey()),
	}

	err = r.apply(c.state)
	if err != nil {
		return nil, fmt.Errorf("unable to apply rotation on state, %w", err)
	}

	return r, nil
}

func getPrivateKey(
	publicKeyHash chore.Hash,
	kvStore *nutsdb.DB,
) (*crypto.PrivateKey, error) {
	tx, err := kvStore.Begin(false)
	if err != nil {
		return nil, fmt.Errorf("unable to begin tx, %w", err)
	}
	// nolint: errcheck
	defer tx.Commit()

	en, err := tx.Get(privateKeysBucket, []byte(publicKeyHash))
	if err != nil {
		return nil, fmt.Errorf("unable to get active key, %w", err)
	}

	pk := &crypto.PrivateKey{}
	err = pk.UnmarshalString(string(en.Value))
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal active key, %w", err)
	}

	return pk, nil
}

func storePrivateKey(
	privateKey crypto.PrivateKey,
	kvStore *nutsdb.DB,
) error {
	tx, err := kvStore.Begin(true)
	if err != nil {
		return fmt.Errorf("unable to begin tx, %w", err)
	}
	// nolint: errcheck
	defer tx.Commit()

	key := getPublicKeyHash(privateKey.PublicKey())
	value, err := privateKey.MarshalString()
	if err != nil {
		return fmt.Errorf("unable to marshal key, %w", err)
	}

	err = tx.Put(privateKeysBucket, []byte(key), []byte(value), 0)
	if err != nil {
		return fmt.Errorf("unable to put key, %w", err)
	}

	return nil
}

func getPublicKeyHash(k crypto.PublicKey) chore.Hash {
	return chore.String(k.String()).Hash()
}
