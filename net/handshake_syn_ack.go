package net

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&HandshakeSynAck{})
}

type HandshakeSynAck struct {
	Nonce     string            `json:"nonce"`
	Signature *crypto.Signature `json:"-"`
}

func (r *HandshakeSynAck) GetType() string {
	return "handshake.syn-ack"
}

func (r *HandshakeSynAck) GetSignature() *crypto.Signature {
	return r.Signature
}

func (r *HandshakeSynAck) SetSignature(s *crypto.Signature) {
	r.Signature = s
}

func (r *HandshakeSynAck) GetAnnotations() map[string]interface{} {
	return map[string]interface{}{}
}

func (r *HandshakeSynAck) SetAnnotations(a map[string]interface{}) {
}
