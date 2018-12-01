package peers

import (
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
)

//go:generate go run nimona.io/go/cmd/objectify -schema /peer -type PeerInfo -out peerinfo_generated.go

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	RawObject    *encoding.Object  `json:"@"`
	Addresses    []string          `json:"addresses"`
	AuthorityKey *crypto.Key       `json:"@authority"`
	SignerKey    *crypto.Key       `json:"@signer"`
	Signature    *crypto.Signature `json:"@sig:O"`
}

// Thumbprint of peer
// TODO rename to ID() or PeerID()?
// TODO(geoah) should this return the authority or the subject's id?
func (pi *PeerInfo) Thumbprint() string {
	return pi.SignerKey.HashBase58()
}

// Address of the peer
func (pi *PeerInfo) Address() string {
	return "peer:" + pi.Thumbprint()
}
