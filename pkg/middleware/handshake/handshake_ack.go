package handshake

import (
	"nimona.io/pkg/crypto"
)

//go:generate go run nimona.io/tools/objectify -schema /handshake.ack -type Ack -in handshake_ack.go -out handshake_ack_generated.go

type Ack struct {
	Nonce     string            `json:"nonce"`
	Signature *crypto.Signature `json:"@signature"`
}
