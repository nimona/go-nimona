package dht

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/encoding"
)

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/provider.request -type ProviderRequest -out provider_request_generated.go

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string `json:"requestID,omitempty"`
	Key       string `json:"key"`

	RawObject *encoding.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@signature"`
}
