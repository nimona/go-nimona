package dht

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/encoding"
	"nimona.io/pkg/peers"
)

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/provider.response -type ProviderResponse -out provider_response_generated.go

type ProviderResponse struct {
	RequestID    string            `json:"requestID,omitempty"`
	Providers    []*Provider       `json:"providers,omitempty"`
	ClosestPeers []*peers.PeerInfo `json:"closestPeers,omitempty"`

	RawObject *encoding.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@signature"`
}
