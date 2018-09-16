package net

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&HandshakeAck{})
}

type HandshakeAck struct {
	Nonce     string            `json:"nonce"`
	Signature *crypto.Signature `json:"-"`
}

func (r *HandshakeAck) GetType() string {
	return "handshake.ack"
}

func (r *HandshakeAck) GetSignature() *crypto.Signature {
	return r.Signature
}

func (r *HandshakeAck) SetSignature(s *crypto.Signature) {
	r.Signature = s
}

func (r *HandshakeAck) GetAnnotations() map[string]interface{} {
	return map[string]interface{}{}
}

func (r *HandshakeAck) SetAnnotations(a map[string]interface{}) {
}
