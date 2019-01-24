package net

import (
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema /handshake.syn -type HandshakeSyn -in handshake_syn.go -out handshake_syn_generated.go

type HandshakeSyn struct {
	RawObject *object.Object `json:"@"`
	Nonce     string         `json:"nonce"`
	PeerInfo  *peer.PeerInfo `json:"peerInfo,omitempty"`
}
