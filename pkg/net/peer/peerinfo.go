package peer

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema /peer -type PeerInfo -in peerinfo.go -out peerinfo_generated.go

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	Addresses    []string          `json:"addresses"`
	Protocols    []string          `json:"protocols"`
	ContentIDs   []string          `json:"contentIDs"`
	ContentTypes []string          `json:"contentTypes"`
	Signature    *crypto.Signature `json:"@signature"`
}

// Fingerprint of signer
func (pi *PeerInfo) Fingerprint() string {
	if pi == nil || pi.Signature == nil || pi.Signature.PublicKey == nil {
		return ""
	}

	return pi.Signature.PublicKey.Fingerprint()
}

// Address of the peer
func (pi *PeerInfo) Address() string {
	return "peer:" + pi.Fingerprint()
}
