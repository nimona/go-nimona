package mesh

import (
	"encoding/json"
	"errors"

	"github.com/keybase/saltpack"
	"github.com/keybase/saltpack/basic"
)

type PeerInfo struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
	PublicKey [32]byte `json:"public_key"`
	Signature []byte   `json:"signature"`
}

type peerInfoClean struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
	PublicKey [32]byte `json:"public_key"`
}

func (pi *PeerInfo) GetPublicKey() basic.PublicKey {
	return basic.PublicKey{
		RawBoxKey: pi.PublicKey,
	}
}

func (pi *PeerInfo) Decrypt(ciphertext string) ([]byte, error) {
	_, raw, _, err := saltpack.Dearmor62DecryptOpen(saltpack.CheckKnownMajorVersion, ciphertext, Keyring)
	return raw, err
}

func (pi *PeerInfo) Marshal() []byte {
	b, _ := json.Marshal(pi)
	return b
}

func (pi *PeerInfo) MarshalWithoutSignature() []byte {
	cpi := &peerInfoClean{
		ID:        pi.ID,
		Addresses: pi.Addresses,
		PublicKey: pi.PublicKey,
	}
	b, _ := json.Marshal(cpi)
	return b
}

func (pi *PeerInfo) IsValid() bool {
	// TODO Fix IsValis
	// pk := DecocdePublicKey(pi.PublicKey)
	// if IDFromPublicKey(*pk) != pi.ID {
	// 	return false
	// }

	// valid, err := Verify(pk, pi.MarshalWithoutSignature(), pi.Signature)
	// if err != nil {
	// 	return false
	// }

	return true
}

func NewPeerInfo(id string, addresses []string, publicKey [32]byte) (*PeerInfo, error) {
	if id == "" {
		return nil, errors.New("missing id")
	}

	if len(addresses) == 0 {
		return nil, errors.New("missing addresses")
	}

	if len(publicKey) == 0 {
		return nil, errors.New("missing public key")
	}

	pi := &PeerInfo{
		ID:        id,
		Addresses: addresses,
		PublicKey: publicKey,
	}

	if !pi.IsValid() {
		return nil, errors.New("id and pk don't match")
	}

	return pi, nil
}
