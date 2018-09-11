package peers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
	"nimona.io/go/log"
)

// LoadOrCreateLocalPeerInfo from/to a JSON encoded file
func (reg *AddressBook) LoadOrCreateLocalPeerInfo(path string) (*PrivatePeerInfo, error) {
	ctx := context.Background()
	if path == "" {
		return nil, errors.New("missing key path")
	}

	// idPath := filepath.Join(path, "identity.json")
	peerPath := filepath.Join(path, "config.json")

	if _, err := os.Stat(peerPath); err == nil {
		return reg.LoadPrivatePeerInfo(peerPath)
	}

	logger := log.Logger(ctx)
	logger.Info("* Configs do not exist, creating new ones.")

	pi, err := reg.CreateNewPeer()
	if err != nil {
		return nil, err
	}

	if err := reg.StorePrivatePeerInfo(pi, peerPath); err != nil {
		return nil, err
	}

	return pi, nil
}

// CreateNewPeer with a new generated key, mostly used for testing
func (reg *AddressBook) CreateNewPeer() (*PrivatePeerInfo, error) {
	peerSigningKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	sk, err := crypto.NewKey(peerSigningKey)
	if err != nil {
		return nil, err
	}

	pi := &PrivatePeerInfo{
		Key:       sk,
		Addresses: []string{},
	}

	return pi, nil
}

type config struct {
	Key string `json:"key"`
}

// LoadPrivatePeerInfo from a JSON encoded file
func (reg *AddressBook) LoadPrivatePeerInfo(path string) (*PrivatePeerInfo, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := config{}
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	fmt.Println(cfg.Key)
	keyi, err := blocks.UnpackDecodeBase58(cfg.Key)
	if err != nil {
		return nil, err
	}

	key := keyi.(*crypto.Key)
	pi := &PrivatePeerInfo{
		Key: key,
	}

	return pi, nil
}

// StorePrivatePeerInfo to a JSON encoded file
func (reg *AddressBook) StorePrivatePeerInfo(pi *PrivatePeerInfo, path string) error {
	ks, err := blocks.PackEncodeBase58(pi.Key)
	if err != nil {
		return err
	}
	cfg := config{
		Key: ks,
	}
	bc, _ := json.MarshalIndent(cfg, "", "  ")
	return ioutil.WriteFile(path, bc, 0644)
}
