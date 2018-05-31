package mesh

import (
	"sync"

	"github.com/keybase/saltpack"

	"github.com/keybase/saltpack/basic"
)

// SecretPeerInfo is a PeerInfo with an additional SecretKey
type SecretPeerInfo struct {
	sync.RWMutex
	PeerInfo
	SecretKey        [32]byte `json:"secret_key"`
	SigningSecretKey [64]byte `json:"signing_secret_key"`
}

// GetSecretKey returns a saltpack SecretKey
func (pi *SecretPeerInfo) GetSecretKey() saltpack.BoxSecretKey {
	return basic.NewSecretKey(&pi.PublicKey, &pi.SecretKey)
}

func (pi *SecretPeerInfo) GetSigningSecretKey() basic.SigningSecretKey {
	return basic.NewSigningSecretKey(&pi.PublicKey, &pi.SigningSecretKey)
}

// UpdateAddresses to allow updating peer addresess
// TODO Consider moving to address book
func (pi *SecretPeerInfo) UpdateAddresses(addresses []string) {
	pi.Lock()
	pi.Addresses = addresses
	pi.Unlock()
}

// ToPeerInfo returns a PeerInfo struct
func (pi *SecretPeerInfo) ToPeerInfo() PeerInfo {
	return PeerInfo{
		ID:        pi.ID,
		Addresses: pi.Addresses,
		PublicKey: pi.PublicKey,
		Signature: pi.Signature,
	}
}
