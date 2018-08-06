package net

import (
	blocks "github.com/nimona/go-nimona/blocks"
)

const (
	TypeHandshake = "handshake"
)

func init() {
	blocks.RegisterContentType(TypeHandshake, HandshakeBlock{})
}

// HandshakeBlock content structure for Handshake content type
type HandshakeBlock struct {
	PeerInfo *blocks.Block `json:"peer_info"`
}
