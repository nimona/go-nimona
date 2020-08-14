package peer

import "nimona.io/pkg/crypto"

// Address of the peer
func (pi *Peer) Address() string {
	if len(pi.Metadata.Signatures) == 0 {
		return ""
	}
	return "peer:" + pi.Metadata.Signatures[0].Signer.String()
}

func (pi *Peer) PublicKey() crypto.PublicKey {
	if len(pi.Metadata.Owners) == 0 {
		return ""
	}
	return pi.Metadata.Owners[0]
}
