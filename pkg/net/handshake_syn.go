package net

import (
	"nimona.io/go/encoding"
	"nimona.io/go/peers"
)

//go:generate go run nimona.io/go/generators/objectify -schema /handshake.syn -type HandshakeSyn -out handshake_syn_generated.go

type HandshakeSyn struct {
	RawObject *encoding.Object `json:"@"`
	Nonce     string           `json:"nonce"`
	PeerInfo  *peers.PeerInfo  `json:"peerInfo,omitempty"`
}
