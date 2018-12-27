package peers

import (
	"nimona.io/pkg/crypto"
)

//go:generate go run nimona.io/tools/objectify -schema /peer.request -type PeerInfoRequest -out peerinfo_request_generated.go

// PeerInfoRequest is a request for a peer info
type PeerInfoRequest struct {
	AuthorityKeyHash string   `json:"authority"`
	SignerKeyHash    string   `json:"signer"`
	Protocols        []string `json:"protocols"`
	ContentIDs       []string `json:"contentIDs"`
	ContentTypes     []string `json:"contentTypes"`

	RequesterAuthorityKey *crypto.Key       `json:"@authority"`
	RequesterSignerKey    *crypto.Key       `json:"@signer"`
	RequestSignature      *crypto.Signature `json:"@signature"`
}
