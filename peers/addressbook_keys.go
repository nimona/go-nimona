package peers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"nimona.io/go/base58"
	"nimona.io/go/log"
	"nimona.io/go/primitives"
)

// loadConfig signing key from/to a JSON encoded file
func (ab *AddressBook) loadConfig(configPath string) error {
	ctx := context.Background()
	peerPath := filepath.Join(configPath, "config.json")
	if _, err := os.Stat(peerPath); err == nil {
		cfg, err := loadConfig(peerPath)
		if err != nil {
			return err
		}
		keyBytes, err := base58.Decode(cfg.Key)
		if err != nil {
			return err
		}
		key := &primitives.Key{}
		keyBlock, err := primitives.Unmarshal(keyBytes)
		if err != nil {
			return err
		}
		key.FromBlock(keyBlock)
		ab.localKey = key
		return nil
	}

	logger := log.Logger(ctx)
	logger.Info("* Configs do not exist, creating new ones.")

	peerSigningKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	localKey, err := primitives.NewKey(peerSigningKey)
	if err != nil {
		return err
	}

	ab.localKey = localKey

	keyBytes, err := primitives.Marshal(localKey.Block())
	if err != nil {
		return err
	}

	cfg := &config{
		Key: base58.Encode(keyBytes),
	}

	if err := storeConfig(cfg, peerPath); err != nil {
		return err
	}

	return nil
}

type config struct {
	Key string `json:"key"`
}

// loadConfig from a JSON encoded file
func loadConfig(path string) (*config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &config{}
	if err := json.Unmarshal(bytes, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// storeConfig to a JSON encoded file
func storeConfig(cfg *config, path string) error {
	bc, _ := json.MarshalIndent(cfg, "", "  ")
	return ioutil.WriteFile(path, bc, 0644)
}
