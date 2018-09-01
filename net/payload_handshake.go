package net

import (
	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/peers"
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
	Signature *blocks.Signature `nimona:",signature" json:"signature"`
	PeerInfo  *peers.PeerInfo   `nimona:"peerInfo" json:"peerInfo"`
}
