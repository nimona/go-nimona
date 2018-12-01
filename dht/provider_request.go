package dht

import (
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
)

//go:generate go run nimona.io/go/cmd/objectify -schema nimona.io/dht/provider.request -type ProviderRequest -out provider_request_generated.go

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string `json:"requestID,omitempty"`
	Key       string `json:"key"`

	RawObject *encoding.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@sig:O"`
}
