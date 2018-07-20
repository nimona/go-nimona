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
func (reg *AddressBook) LoadOrCreateLocalPeerInfo(path string) (*PrivatePeerInfo, error) {
	if path == "" {
		return nil, errors.New("missing key path")
	}

	idPath := filepath.Join(path, "identity.json")
	peerPath := filepath.Join(path, "peer.json")

	if _, err := os.Stat(peerPath); err == nil {
		return reg.LoadPrivatePeerInfo(peerPath)
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

	if err := reg.StorePrivatePeerInfo(pi, peerPath); err != nil {
		return nil, err
	}

	return pi, nil
}

// CreateNewPeer with a new generated key, mostly used for testing
func (reg *AddressBook) CreateNewPeer() (*PrivatePeerInfo, error) {
	peerSigningKey, err := btckey.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pi := &PrivatePeerInfo{
		ID:         fmt.Sprintf("01x%x", peerSigningKey.PublicKey.ToBytes()),
		PrivateKey: fmt.Sprintf("01x%x", peerSigningKey.ToBytes()),
	}

	return pi, nil
}

// LoadPrivatePeerInfo from a JSON encoded file
func (reg *AddressBook) LoadPrivatePeerInfo(path string) (*PrivatePeerInfo, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pi := &PrivatePeerInfo{}
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

// StorePrivatePeerInfo to a JSON encoded file
func (reg *AddressBook) StorePrivatePeerInfo(pi *PrivatePeerInfo, path string) error {
	raw, err := json.MarshalIndent(pi, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, raw, 0644)
}

// Sign data given private key in its prefixed and compressed format
func Sign(data []byte, privateKey string) ([]byte, error) {
	// TODO check private key format
	return btckey.Sign(data, privateKey[3:])
}

// Verify the signature of some data given a public key in its prefixed and
// compressed format
func Verify(publicKey string, data, signature []byte) error {
	digest := sha256.Sum256(data)
	publicKeyBytes, _ := hex.DecodeString(publicKey[3:])
	ok := btckey.Verify(publicKeyBytes, signature, digest[:])
	if !ok {
		return errors.New("could not verify signature")
	}
	return nil
}
