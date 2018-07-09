package peer

import (
	"errors"
	"time"

	"github.com/keybase/saltpack"
	"github.com/keybase/saltpack/basic"
)

// PeerInfo holds the information wire needs to connect to a remote peer
type PeerInfo struct {
	ID              string    `json:"id"`
	Version         int       `json:"version"`
	Addresses       []string  `json:"addresses"`
	PublicKey       [32]byte  `json:"public_key"`
	Signature       []byte    `json:"signature"`
	CreatedAt       time.Time `json:"create_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LastConnectedAt time.Time `json:"last_connected_at"`
}

type peerInfoClean struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
	PublicKey [32]byte `json:"public_key"`
}

// GetPublicKey returns the public key of the peer as a BoxPublicKey
func (pi *PeerInfo) GetPublicKey() saltpack.BoxPublicKey {
	return basic.PublicKey{
		RawBoxKey: pi.PublicKey,
	}
}

// IsValid checks if the signature is valid
func (pi *PeerInfo) IsValid() bool {
	// TODO Implement
	return true
}

// NewPeerInfo from an id, an address, and a public key
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if !pi.IsValid() {
		return nil, errors.New("id and pk don't match")
	}

	return pi, nil
}
