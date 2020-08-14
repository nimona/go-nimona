package peer

import "nimona.io/pkg/crypto"

// Address of the peer
func (pi *Peer) Address() string {
	if pi.Metadata.Signature.Signer.IsEmpty() {
		return ""
	}
	return "peer:" + pi.Metadata.Owner.String()
}

func (pi *Peer) PublicKey() crypto.PublicKey {
	return pi.Metadata.Owner
}
