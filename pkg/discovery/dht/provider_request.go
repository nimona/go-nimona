package dht

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/provider.request -type ProviderRequest -in provider_request.go -out provider_request_generated.go

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string `json:"requestID,omitempty"`
	Key       string `json:"key"`

	RawObject *object.Object    `json:"@"`
	Signer    *crypto.PublicKey `json:"@signer"`
	Authority *crypto.PublicKey `json:"@authority"`
	Signature *crypto.Signature `json:"@signature"`
}
