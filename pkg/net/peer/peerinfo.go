package peer

import (
	"fmt"

	"nimona.io/pkg/crypto"
)

//go:generate go run nimona.io/tools/objectify -schema /peer -type PeerInfo -in peerinfo.go -out peerinfo_generated.go

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	Addresses    []string          `json:"addresses"`
	Protocols    []string          `json:"protocols"`
	ContentIDs   []string          `json:"contentIDs"`
	ContentTypes []string          `json:"contentTypes"`
	SignerKey    *crypto.PublicKey `json:"@signer"`
	Signature    *crypto.Signature `json:"@signature"`
}

// HashBase58 of peer
// TODO rename to ID() or PeerIDs()?
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
	ppub := p.SignerKey.Hash
	return fmt.Sprintf(
		"(ppub: %s; addrs: %v)",
		ppub,
		p.Addresses,
	)
}
