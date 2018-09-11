package net

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
	"nimona.io/go/peers"
)

func init() {
	blocks.RegisterContentType(&HandshakePayload{})
}

// HandshakePayload content structure for Handshake content type
type HandshakePayload struct {
	PeerInfo  *peers.PeerInfo   `json:"peerInfo"`
	Signature *crypto.Signature `json:"-"`
}

func (r *HandshakePayload) GetType() string {
	return "handshake"
}

func (r *HandshakePayload) GetSignature() *crypto.Signature {
	return r.Signature
}

func (r *HandshakePayload) SetSignature(s *crypto.Signature) {
	r.Signature = s
}

func (r *HandshakePayload) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (r *HandshakePayload) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
