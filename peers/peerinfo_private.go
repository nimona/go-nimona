package peers

import (
	"crypto/ecdsa"

	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

// PrivatePeerInfo is a PeerInfo with an additional PrivateKey
type PrivatePeerInfo struct {
	Key       *crypto.Key `json:"key"`
	Addresses []string    `json:"-"`
}

func (pi *PrivatePeerInfo) Thumbprint() string {
	return pi.GetPeerInfo().Thumbprint()
}

func (pi *PrivatePeerInfo) GetPeerInfo() *PeerInfo {
	ppi := &PeerInfo{
		Addresses: pi.Addresses,
	}

	sig, err := blocks.Sign(ppi, pi.Key)
	if err != nil {
		panic(err)
	}

	ppi.Signature = sig
	return ppi
}

// GetPrivateKey returns the private key
func (pi *PrivatePeerInfo) GetPrivateKey() *crypto.Key {
	return pi.Key
}

// GetPublicKey returns the public key
func (pi *PrivatePeerInfo) GetPublicKey() *crypto.Key {
	pk := pi.Key.Materialize().(*ecdsa.PrivateKey).Public().(*ecdsa.PublicKey)
	bpk, err := crypto.NewKey(pk)
	if err != nil {
		panic(err)
	}
	return bpk
}
