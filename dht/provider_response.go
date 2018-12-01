package dht

import (
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/peers"
)

//go:generate go run nimona.io/go/cmd/objectify -schema nimona.io/dht/provider.response -type ProviderResponse -out provider_response_generated.go

type ProviderResponse struct {
	RequestID    string            `json:"requestID,omitempty"`
	Providers    []*Provider       `json:"providers,omitempty"`
	ClosestPeers []*peers.PeerInfo `json:"closestPeers,omitempty"`

	RawObject *encoding.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@sig:O"`
}
