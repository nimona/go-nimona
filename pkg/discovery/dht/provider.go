package dht

import (
	"nimona.io/pkg/crypto"
)

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/provider -type Provider -in provider.go -out provider_generated.go

// Provider payload
type Provider struct {
	ObjectIDs []string          `json:"objectIDs"`
	Signature *crypto.Signature `json:"@signature"`
}
