package mesh

import (
	"encoding/json"
	"errors"
	"time"
)

type Status int

const (
	Connected Status = iota
	NotConnected
	CanConnect
	ErrConnecting
)

type PeerInfo struct {
	ID              string    `json:"id"`
	Addresses       []string  `json:"addresses"`
	PublicKey       []byte    `json:"public_key"`
	Signature       []byte    `json:"signature"`
	Status          Status    `json:"status"`
	CreatedAt       time.Time `json:"create_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LastConnectedAt time.Time `json:"last_connected_at"`
}
type peerInfoClean struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
	PublicKey []byte   `json:"public_key"`
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
	pk := DecocdePublicKey(pi.PublicKey)
	if IDFromPublicKey(*pk) != pi.ID {
		return false
	}

	valid, err := Verify(pk, pi.MarshalWithoutSignature(), pi.Signature)
	if err != nil {
		return false
	}

	return valid
}

func NewPeerInfo(id string, addresses []string, publicKey []byte) (*PeerInfo, error) {
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
