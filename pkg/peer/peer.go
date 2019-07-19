package peer

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema nimona.io/discovery/peer -type Peer -in peer.go -out peer_generated.go

// Peer holds the information exchange needs to connect to a remote peer
type Peer struct {
	Addresses    []string          `json:"addresses:as"`
	ContentTypes []string          `json:"contentTypes:as"`
	Signature    *crypto.Signature `json:"@signature:o"`
}

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

func (pi Peer) Bloom() []int {
	return []int{}
}
