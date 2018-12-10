package net

import (
	"nimona.io/go/crypto"
)

//go:generate go run nimona.io/go/cmd/objectify -schema /handshake.ack -type HandshakeAck -out handshake_ack_generated.go

type HandshakeAck struct {
	Nonce     string            `json:"nonce"`
	Signer    *crypto.Key       `json:"@signer"`
	Signature *crypto.Signature `json:"@signature"`
}
