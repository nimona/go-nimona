package mesh

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/keybase/saltpack/basic"

	"golang.org/x/crypto/nacl/box"
)

var Keyring = basic.NewKeyring()

// LoadOrCreateLocalPeerInfo from/to a JSON encoded file
func LoadOrCreateLocalPeerInfo(path string) (*SecretPeerInfo, error) {
	if path == "" {
		return nil, errors.New("missing key path")
	}

	if _, err := os.Stat(path); err == nil {
		return LoadSecretPeerInfo(path)
	}

	log.Printf("* Key path does not exist, creating new key in '%s'\n", path)

	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pi := &SecretPeerInfo{
		PeerInfo: PeerInfo{
			Addresses: []string{},
			PublicKey: *pub,
		},
		SecretKey: *priv,
	}

	pi.ID = fmt.Sprintf("%x", pi.GetPublicKey().ToKID())

	Keyring.ImportBoxKey(pub, priv)

	if err := StoreSecretPeerInfo(pi, path); err != nil {
		return nil, err
	}

	return pi, nil
}

// LoadSecretPeerInfo from a JSON encoded file
func LoadSecretPeerInfo(path string) (*SecretPeerInfo, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pi := &SecretPeerInfo{}
	if err := json.Unmarshal(raw, &pi); err != nil {
		return nil, err
	}

	Keyring.ImportBoxKey(&pi.PublicKey, &pi.SecretKey)

	return pi, nil
}

// StoreSecretPeerInfo to a JSON encoded file
func StoreSecretPeerInfo(pi *SecretPeerInfo, path string) error {
	raw, err := json.Marshal(pi)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, raw, 0644)
}
