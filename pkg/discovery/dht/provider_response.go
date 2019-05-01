package dht

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/provider.response -type ProviderResponse -in provider_response.go -out provider_response_generated.go

type ProviderResponse struct {
	RequestID    string           `json:"requestID,omitempty"`
	Providers    []*Provider      `json:"providers,omitempty"`
	ClosestPeers []*peer.PeerInfo `json:"closestPeers,omitempty"`

	RawObject *object.Object    `json:"@"`
	Signer    *crypto.PublicKey `json:"@signer"`
	Signature *crypto.Signature `json:"@signature"`
}
