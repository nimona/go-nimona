package net

import (
	blocks "github.com/nimona/go-nimona/blocks"
)

const (
	// TypeHandshake is the type of HandshakePayload Block
	TypeHandshake = "handshake"
)

func init() {
	blocks.RegisterContentType(TypeHandshake, HandshakePayload{})
}

// HandshakePayload content structure for Handshake content type
type HandshakePayload struct {
	PeerInfo *blocks.Block `json:"peer_info"`
}
