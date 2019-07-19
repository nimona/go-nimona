package handshake

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema /handshake.ack -type Ack -in handshake_ack.go -out handshake_ack_generated.go

type Ack struct {
	Nonce     string            `json:"nonce:s"`
	Signature *crypto.Signature `json:"@signature:o"`
}
