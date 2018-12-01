package dht

import (
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
)

//go:generate go run nimona.io/go/cmd/objectify -schema nimona.io/dht/provider -type Provider -out provider_generated.go

// Provider payload
type Provider struct {
	BlockIDs []string `json:"blockIDs"`

	RawObject *encoding.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@sig:O"`
}
