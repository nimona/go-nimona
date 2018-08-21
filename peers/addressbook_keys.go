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
	"github.com/nimona/go-nimona/keys"
)

// LoadOrCreateLocalPeerInfo from/to a JSON encoded file
func (reg *AddressBook) LoadOrCreateLocalPeerInfo(path string) (*PrivatePeerInfo, error) {
	if path == "" {
		return nil, errors.New("missing key path")
	}

	// idPath := filepath.Join(path, "identity.json")
	peerPath := filepath.Join(path, "peer.json")

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

	sk, err := keys.New(peerSigningKey)
	if err != nil {
		return nil, err
	}

	msk, err := sk.Marshal()
	if err != nil {
		return nil, err
	}

	pk, err := keys.New(&peerSigningKey.PublicKey)
	mpk, err := pk.Marshal()
	if err != nil {
		return nil, err
	}

	pi := &PrivatePeerInfo{
		ID:         blocks.Base58Encode(mpk),
		PrivateKey: blocks.Base58Encode(msk),
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
