package peer

import (
	"github.com/keybase/saltpack"

	"github.com/keybase/saltpack/basic"
)

// SecretPeerInfo is a PeerInfo with an additional SecretKey
type SecretPeerInfo struct {
	PeerInfo
	SecretKey [32]byte `json:"secret_key"`
}

// GetSecretKey returns a saltpack SecretKey
func (pi *SecretPeerInfo) GetSecretKey() saltpack.BoxSecretKey {
	return basic.NewSecretKey(&pi.PublicKey, &pi.SecretKey)
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
