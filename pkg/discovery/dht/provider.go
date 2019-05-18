package dht

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema nimona.io/dht/provider -type Provider -in provider.go -out provider_generated.go

// Provider payload
type Provider struct {
	ObjectIDs []string          `json:"objectIDs"`
	Signature *crypto.Signature `json:"@signature"`
}
