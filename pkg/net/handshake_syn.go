package net

import (
	"nimona.io/pkg/encoding"
	"nimona.io/pkg/peers"
)

//go:generate go run nimona.io/tools/objectify -schema /handshake.syn -type HandshakeSyn -in handshake_syn.go -out handshake_syn_generated.go

type HandshakeSyn struct {
	RawObject *encoding.Object `json:"@"`
	Nonce     string           `json:"nonce"`
	PeerInfo  *peers.PeerInfo  `json:"peerInfo,omitempty"`
}
