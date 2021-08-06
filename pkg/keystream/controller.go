package keystream

import (
	"fmt"
	"sync"

	"github.com/xujiajun/nutsdb"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
)

const (
	// bucketPrivateKeys is the bucket that holds private keys
	bucketPrivateKeys = "keystream_private_keys"
	// bucketConfigs is the bucket that holds config key value pairs
	bucketConfigs = "keystream_configs"
	// keyKeyStreamRootHash is the config key for the keystream's root
	keyKeyStreamRootHash = "keystream_root_hash"
)

type (
	// Controller deals with the key management and event transitions for a
	// single key stream
	Controller struct {
		mutex             sync.RWMutex
		kvStore           *nutsdb.DB
		objectStore       objectstore.Store
		state             *KeyStream
		currentPrivateKey crypto.PrivateKey
		newKey            func() (crypto.PrivateKey, error)
	}
)

func NewController(
	kvStore *nutsdb.DB,
	objectStore objectstore.Store,
) (*Controller, error) {
	var keyStream *KeyStream
	keyStreamRootHashBytes, err := getConfigValue(keyKeyStreamRootHash, kvStore)
	if err == nil {
		eventStream, err := objectStore.GetByStream(
			chore.Hash(keyStreamRootHashBytes),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get keystream objects, %w", err)
		}
		keyStream, err = FromStream(eventStream)
		if err != nil {
			return nil, fmt.Errorf("unable to create state, %w", err)
		}
	}

	if keyStream == nil {
		k0, err := crypto.NewEd25519PrivateKey()
		if err != nil {
			return nil, fmt.Errorf("unable to generate key, %w", err)
		}
		k1, err := crypto.NewEd25519PrivateKey()
		if err != nil {
			return nil, fmt.Errorf("unable to generate key, %w", err)
		}
		err = putPrivateKey(k0, kvStore)
		if err != nil {
			return nil, fmt.Errorf("unable to put key, %w", err)
		}
		err = putPrivateKey(k1, kvStore)
		if err != nil {
			return nil, fmt.Errorf("unable to put key, %w", err)
		}
		inceptionEvent := &Inception{
			Metadata: object.Metadata{
				Sequence: 0,
			},
			Version:       Version,
			Key:           k0.PublicKey(),
			NextKeyDigest: getPublicKeyHash(k1.PublicKey()),
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
		err = putConfigValue(
			keyKeyStreamRootHash,
			[]byte(inceptionObject.Hash()),
			kvStore,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to put config value, %w", err)
		}
		keyStream, err = FromStream(
			object.NewReadCloserFromObjects(
				[]*object.Object{
					inceptionObject,
				},
			),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create new state, %w", err)
		}
	}

	// TODO check that we have the next key as well

	pk, err := getPrivateKey(getPublicKeyHash(keyStream.ActiveKey), kvStore)
	if err != nil {
		return nil, fmt.Errorf("unable to get private active key, %w", err)
	}

	c := &Controller{
		mutex:             sync.RWMutex{},
		kvStore:           kvStore,
		objectStore:       objectStore,
		state:             keyStream,
		currentPrivateKey: *pk,
		newKey:            crypto.NewEd25519PrivateKey,
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

	newCurrentKey, err := getPrivateKey(c.state.NextKeyDigest, c.kvStore)
	if err != nil {
		return nil, fmt.Errorf("unable to get next private key, %w", err)
	}

	err = putPrivateKey(newNextKey, c.kvStore)
	if err != nil {
		return nil, fmt.Errorf("unable to store new next private key, %w", err)
	}

	r := &Rotation{
		Metadata:      object.Metadata{},
		Version:       Version,
		Key:           newCurrentKey.PublicKey(),
		NextKeyDigest: getPublicKeyHash(newNextKey.PublicKey()),
	}

	c.currentPrivateKey = *newCurrentKey

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

	en, err := tx.Get(bucketPrivateKeys, []byte(publicKeyHash))
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

func putPrivateKey(
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

	err = tx.Put(bucketPrivateKeys, []byte(key), []byte(value), 0)
	if err != nil {
		return fmt.Errorf("unable to put key, %w", err)
	}

	return nil
}

func getPublicKeyHash(k crypto.PublicKey) chore.Hash {
	return chore.String(k.String()).Hash()
}

func getConfigValue(
	key string,
	kvStore *nutsdb.DB,
) ([]byte, error) {
	tx, err := kvStore.Begin(false)
	if err != nil {
		return nil, fmt.Errorf("unable to begin tx, %w", err)
	}
	// nolint: errcheck
	defer tx.Commit()

	res, err := tx.Get(bucketConfigs, []byte(keyKeyStreamRootHash))
	if err != nil {
		return nil, fmt.Errorf("unable to get active key, %w", err)
	}

	return res.Value, nil
}

func putConfigValue(
	key string,
	value []byte,
	kvStore *nutsdb.DB,
) error {
	tx, err := kvStore.Begin(true)
	if err != nil {
		return fmt.Errorf("unable to begin tx, %w", err)
	}
	// nolint: errcheck
	defer tx.Commit()

	return tx.Put(bucketConfigs, []byte(keyKeyStreamRootHash), value, 0)
}
