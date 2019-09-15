package peer

import (
	"nimona.io/pkg/crypto"
)

// Fingerprint of signer
func (pi *Peer) Fingerprint() crypto.Fingerprint {
	if pi == nil || pi.Signature == nil || pi.Signature.PublicKey == nil {
		return ""
	}

	return pi.Signature.PublicKey.Fingerprint()
}

// Address of the peer
func (pi *Peer) Address() string {
	return "peer:" + pi.Fingerprint().String()
}
