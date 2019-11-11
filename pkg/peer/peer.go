package peer

import "nimona.io/pkg/crypto"

// Address of the peer
func (pi *Peer) Address() string {
	return "peer:" + pi.Signature.Signer.Subject.String()
}

func (pi *Peer) PublicKey() crypto.PublicKey {
	return pi.Signature.Signer.Subject
}
