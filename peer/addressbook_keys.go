package peer

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// LoadOrCreateLocalPeerInfo from/to a JSON encoded file
func (reg *AddressBook) LoadOrCreateLocalPeerInfo(path string) (*SecretPeerInfo, error) {
	if path == "" {
		return nil, errors.New("missing key path")
	}

	if _, err := os.Stat(path); err == nil {
		return reg.LoadSecretPeerInfo(path)
	}

	log.Printf("* Key path does not exist, creating new key in '%s'\n", path)

	signingKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	privateKeyBytes, _ := x509.MarshalECPrivateKey(signingKey)
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&signingKey.PublicKey)

	id := &LocalIdentity{
		ID:         fmt.Sprintf("N0x%x", publicKeyBytes),
		Version:    0,
		PrivateKey: privateKeyBytes,
		PublicKey:  publicKeyBytes,
		Peers:      &PeerInfoCollection{},
	}

	idBytes, _ := json.MarshalIndent(id, "", "  ")
	fmt.Println(string(idBytes))

	pi, err := reg.CreateNewPeer()
	if err != nil {
		return nil, err
	}

	if err := reg.StoreSecretPeerInfo(pi, path); err != nil {
		return nil, err
	}

	return pi, nil
}

// CreateNewPeer with a new generated key, mostly used for testing
func (reg *AddressBook) CreateNewPeer() (*SecretPeerInfo, error) {
	peerPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	peerPrivateKeyBytes := x509.MarshalPKCS1PrivateKey(peerPrivateKey)
	peerPublicKeyBytes := x509.MarshalPKCS1PublicKey(&peerPrivateKey.PublicKey)

	pi := &SecretPeerInfo{
		PeerInfo: PeerInfo{
			ID:        fmt.Sprintf("P0x%x", peerPublicKeyBytes),
			Addresses: []string{},
			PublicKey: peerPublicKeyBytes,
		},
		SecretKey: peerPrivateKeyBytes,
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

// StoreSecretPeerInfo to a JSON encoded file
func (reg *AddressBook) StoreSecretPeerInfo(pi *SecretPeerInfo, path string) error {
	raw, err := json.Marshal(pi)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, raw, 0644)
}
