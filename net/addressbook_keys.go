package net

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/apisit/btckeygenie/btckey"
)

// LoadOrCreateLocalPeerInfo from/to a JSON encoded file
func (reg *AddressBook) LoadOrCreateLocalPeerInfo(path string) (*SecretPeerInfo, error) {
	if path == "" {
		return nil, errors.New("missing key path")
	}

	idPath := filepath.Join(path, "identity.json")
	peerPath := filepath.Join(path, "peer.json")

	if _, err := os.Stat(peerPath); err == nil {
		return reg.LoadSecretPeerInfo(peerPath)
	}

	log.Printf("* Configs do not exist, creating new ones.")

	signingKey, err := btckey.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	id := &PrivateIdentity{
		ID:         fmt.Sprintf("00x%x", signingKey.PublicKey.ToBytes()),
		PrivateKey: fmt.Sprintf("00x%x", signingKey.ToBytes()),
		Peers:      &PeerInfoCollection{},
	}

	if err := reg.StorePrivateIdentity(id, idPath); err != nil {
		return nil, err
	}

	pi, err := reg.CreateNewPeer()
	if err != nil {
		return nil, err
	}

	if err := reg.StoreSecretPeerInfo(pi, peerPath); err != nil {
		return nil, err
	}

	return pi, nil
}

// CreateNewPeer with a new generated key, mostly used for testing
func (reg *AddressBook) CreateNewPeer() (*SecretPeerInfo, error) {
	peerSigningKey, err := btckey.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pi := &SecretPeerInfo{
		ID:         fmt.Sprintf("01x%x", peerSigningKey.PublicKey.ToBytes()),
		PrivateKey: fmt.Sprintf("01x%x", peerSigningKey.ToBytes()),
	}

	return pi, nil
}

// LoadSecretPeerInfo from a JSON encoded file
func (reg *AddressBook) LoadSecretPeerInfo(path string) (*SecretPeerInfo, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pi := &SecretPeerInfo{}
	if err := json.Unmarshal(raw, &pi); err != nil {
		return nil, err
	}

	return pi, nil
}

// StorePrivateIdentity to a JSON encoded file
func (reg *AddressBook) StorePrivateIdentity(pi *PrivateIdentity, path string) error {
	raw, err := json.MarshalIndent(pi, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, raw, 0644)
}

// StoreSecretPeerInfo to a JSON encoded file
func (reg *AddressBook) StoreSecretPeerInfo(pi *SecretPeerInfo, path string) error {
	raw, err := json.MarshalIndent(pi, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, raw, 0644)
}

func Sign(data []byte, privateKey string) ([]byte, error) {
	return btckey.Sign(data, privateKey[3:])
}

func Verify(id string, data, signature []byte) error {
	digest := sha256.Sum256(data)
	publicKeyBytes, _ := hex.DecodeString(id[3:])
	publicKey := btckey.PublicKey{}
	publicKey.FromBytes(publicKeyBytes)
	ok := btckey.Verify(publicKeyBytes, signature, digest[:])
	if !ok {
		return errors.New("could not verify signature")
	}
	return nil
}
