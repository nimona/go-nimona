package peer

import "nimona.io/pkg/crypto"

// Address of the peer
func (pi *Peer) Address() string {
	return "peer:" + pi.Header.Signature.Signer.String()
}

func (pi *Peer) PublicKey() crypto.PublicKey {
	return pi.Header.Signature.Signer
}
