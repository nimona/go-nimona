package net

import (
	"nimona.io/pkg/crypto"
)

//go:generate go run nimona.io/tools/objectify -schema /handshake.ack -type HandshakeAck -out handshake_ack_generated.go

type HandshakeAck struct {
	Nonce     string            `json:"nonce"`
	Signer    *crypto.Key       `json:"@signer"`
	Signature *crypto.Signature `json:"@signature"`
}
