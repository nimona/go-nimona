package peers

import (
	"nimona.io/go/crypto"
)

//go:generate go run nimona.io/go/generators/objectify -schema /peer -type PeerInfo -out peerinfo_generated.go

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	Addresses    []string          `json:"addresses"`
	Protocols    []string          `json:"protocols"`
	ContentIDs   []string          `json:"contentIDs"`
	ContentTypes []string          `json:"contentTypes"`
	AuthorityKey *crypto.Key       `json:"@authority"`
	SignerKey    *crypto.Key       `json:"@signer"`
	Signature    *crypto.Signature `json:"@signature"`
}

// HashBase58 of peer
// TODO rename to ID() or PeerID()?
// TODO(geoah) should this return the authority or the subject's id?
func (pi *PeerInfo) HashBase58() string {
	return pi.SignerKey.HashBase58()
}

// Address of the peer
func (pi *PeerInfo) Address() string {
	return "peer:" + pi.HashBase58()
}
