package net

import (
	"nimona.io/go/encoding"
	"nimona.io/go/peers"
)

//go:generate go run nimona.io/go/cmd/objectify -schema /handshake.syn-ack -type HandshakeSynAck -out handshake_syn_ack_generated.go

type HandshakeSynAck struct {
	RawObject *encoding.Object `json:"@"`
	Nonce     string           `json:"nonce"`
	PeerInfo  *peers.PeerInfo  `json:"peerInfo,omitempty"`
}
