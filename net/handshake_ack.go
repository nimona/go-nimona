package net

import (
	"nimona.io/go/encoding"
)

//go:generate go run nimona.io/go/cmd/objectify -schema /handshake.ack -type HandshakeAck -out handshake_ack_generated.go

type HandshakeAck struct {
	RawObject *encoding.Object `json:"@"`
	Nonce     string           `json:"nonce"`
}
