package peer

import (
	"crypto/rsa"
	"crypto/x509"
)

// SecretPeerInfo is a PeerInfo with an additional SecretKey
type SecretPeerInfo struct {
	PeerInfo
	SecretKey []byte `json:"secret_key"`
}

// GetSecretKey returns a saltpack SecretKey
func (pi *SecretPeerInfo) GetSecretKey() *rsa.PrivateKey {
	privateKey, err := x509.ParsePKCS1PrivateKey(pi.SecretKey)
	if err != nil {
		panic("could not get secret for local peer, err=" + err.Error())
	}
	return privateKey
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
