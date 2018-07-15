package net

import (
	"crypto/rsa"
	"crypto/x509"
)

// SecretPeerInfo is a PeerInfo with an additional PrivateKey
type SecretPeerInfo struct {
	PeerInfo
	PrivateKey []byte `json:"secret_key"`
}

// GetPrivateKey returns a saltpack PrivateKey
func (pi *SecretPeerInfo) GetPrivateKey() *rsa.PrivateKey {
	privateKey, err := x509.ParsePKCS1PrivateKey(pi.PrivateKey)
	if err != nil {
		panic("could not get secret for local peer, err=" + err.Error())
	}
	return privateKey
}

// Message returns a PeerInfo struct
func (pi *SecretPeerInfo) Message() PeerInfo {
	return PeerInfo{
		ID:        pi.ID,
		Addresses: pi.Addresses,
		PublicKey: pi.PublicKey,
		Signature: pi.Signature,
	}
}
