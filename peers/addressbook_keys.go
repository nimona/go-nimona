package peers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"

	"github.com/apisit/rfc6979"
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

// SignData given some bytes and a private key in its prefixed and compressed format
func SignData(data []byte, key keys.Key) ([]byte, error) {
	mKey, err := key.Materialize()
	if err != nil {
		return nil, err
	}

	pKey, ok := mKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("only ecdsa keys are currently supported")
	}

	digest := sha256.Sum256(data)
	r, s, err := rfc6979.SignECDSA(pKey, digest[:], sha256.New)
	if err != nil {
		return nil, err
	}

	// TODO replace signature with proper Block-based signature.
	params := pKey.Curve.Params()
	curveOrderByteSize := params.P.BitLen() / 8
	rBytes, sBytes := r.Bytes(), s.Bytes()
	signature := make([]byte, curveOrderByteSize*2)
	copy(signature[curveOrderByteSize-len(rBytes):], rBytes)
	copy(signature[curveOrderByteSize*2-len(sBytes):], sBytes)

	return signature, nil
}

// Verify the signature of some data given a public key in its prefixed and
// compressed format
func Verify(key keys.Key, data, signature []byte) error {
	mKey, err := key.Materialize()
	if err != nil {
		return err
	}

	pKey, ok := mKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("only ecdsa keys are currently supported")
	}

	digest := sha256.Sum256(data)
	rBytes := new(big.Int).SetBytes(signature[0:32])
	sBytes := new(big.Int).SetBytes(signature[32:64])
	if ok := ecdsa.Verify(pKey, digest[:], rBytes, sBytes); !ok {
		return errors.New("could not verify signature")
	}

	return nil
}
