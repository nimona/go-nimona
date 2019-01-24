package crypto

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"nimona.io/internal/log"
	"nimona.io/pkg/object"
)

// LoadKey returns a key from a file if it exists, else will create it
func LoadKey(keyPath string) (*Key, error) {
	if _, err := os.Stat(keyPath); err == nil {
		bytes, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, errors.Wrap(err, "could not read key file")
		}

		o, err := object.Unmarshal(bytes)
		if err != nil {
			return nil, errors.Wrap(err, "could not unmarshal key file")
		}

		key := &Key{}
		if err := key.FromObject(o); err != nil {
			return nil, errors.Wrap(err, "could not convert object to key")
		}

		if _, ok := key.Materialize().(*ecdsa.PrivateKey); !ok {
			return nil, errors.New("invalid key type, only ecdsa keys are allowed")
		}

		return key, nil
	}

	ctx := context.Background()
	logger := log.Logger(ctx)
	logger.Info("* Configs do not exist, creating new ones.")

	peerSigningKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	key, err := NewKey(peerSigningKey)
	if err != nil {
		return nil, err
	}

	keyBytes, err := object.Marshal(key.ToObject())
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(keyPath, keyBytes, 0644); err != nil {
		return nil, err
	}

	return key, nil
}
