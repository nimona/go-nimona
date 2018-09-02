package peers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/nimona/go-nimona/blocks"
)

// LoadOrCreateLocalPeerInfo from/to a JSON encoded file
func (reg *AddressBook) LoadOrCreateLocalPeerInfo(path string) (*PrivatePeerInfo, error) {
	if path == "" {
		return nil, errors.New("missing key path")
	}

	// idPath := filepath.Join(path, "identity.json")
	peerPath := filepath.Join(path, "config.json")

	if _, err := os.Stat(peerPath); err == nil {
		return reg.LoadPrivatePeerInfo(peerPath)
	}

	log.Printf("* Configs do not exist, creating new ones.")

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

	sk, err := blocks.NewKey(peerSigningKey)
	if err != nil {
		return nil, err
	}

	// msk, err := sk.MarshalBinary()
	// if err != nil {
	// 	return nil, err
	// }

	// pk, err := blocks.New(&peerSigningKey.PublicKey)
	// mpk, err := pk.MarshalBinary()
	// if err != nil {
	// 	return nil, err
	// }

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

	keyBytes, err := blocks.Base58Decode(cfg.Key)
	if err != nil {
		return nil, err
	}
	keyi, err := blocks.Unmarshal(keyBytes)
	if err != nil {
		return nil, err
	}

	key := keyi.(*blocks.Key)
	pi := &PrivatePeerInfo{
		Key: key,
	}

	return pi, nil
}

// StorePrivatePeerInfo to a JSON encoded file
func (reg *AddressBook) StorePrivatePeerInfo(pi *PrivatePeerInfo, path string) error {
	key, _ := blocks.MarshalBase58(*pi.Key)
	cfg := config{Key: key}
	bc, _ := json.MarshalIndent(cfg, "", "  ")
	return ioutil.WriteFile(path, bc, 0644)
}
