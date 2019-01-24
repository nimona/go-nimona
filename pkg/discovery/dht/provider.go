package dht

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/provider -type Provider -in provider.go -out provider_generated.go

// Provider payload
type Provider struct {
	BlockIDs []string `json:"blockIDs"`

	RawObject *object.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@signature"`
}
