package peer

import (
	"fmt"

	"nimona.io/pkg/crypto"
)

//go:generate go run nimona.io/tools/objectify -schema /peer -type PeerInfo -in peerinfo.go -out peerinfo_generated.go

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	Addresses    []string `json:"addresses"`
	Protocols    []string `json:"protocols"`
	ContentIDs   []string `json:"contentIDs"`
	ContentTypes []string `json:"contentTypes"`
	// AuthorityKey *crypto.Key       `json:"@authority"`
	SignerKey *crypto.PublicKey `json:"@signer"`
	Signature *crypto.Signature `json:"@signature"`
	// Mandate      *crypto.Mandate   `json:"@mandate"`
}

// HashBase58 of peer
// TODO rename to ID() or PeerID()?
// TODO(geoah) should this return the authority or the subject's id?
func (pi *PeerInfo) HashBase58() string {
	if pi == nil {
		return ""
	}
	return pi.SignerKey.Hash
}

// Address of the peer
func (pi *PeerInfo) Address() string {
	return "peer:" + pi.HashBase58()
}

// String to allow pretty printing peers
func (p *PeerInfo) String() string {
	apub := ""
	// if p.AuthorityKey != nil {
	// 	apub = p.AuthorityKey.HashBase58()
	// }
	ppub := p.SignerKey.Hash
	return fmt.Sprintf(
		"(apub: %s; ppub: %s; addrs: %v)",
		apub,
		ppub,
		p.Addresses,
	)
}
