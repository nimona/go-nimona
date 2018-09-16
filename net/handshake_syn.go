package net

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&HandshakeSyn{})
}

type HandshakeSyn struct {
	Nonce     string            `json:"nonce"`
	Signature *crypto.Signature `json:"-"`
}

func (r *HandshakeSyn) GetType() string {
	return "handshake.syn"
}

func (r *HandshakeSyn) GetSignature() *crypto.Signature {
	return r.Signature
}

func (r *HandshakeSyn) SetSignature(s *crypto.Signature) {
	r.Signature = s
}

func (r *HandshakeSyn) GetAnnotations() map[string]interface{} {
	return map[string]interface{}{}
}

func (r *HandshakeSyn) SetAnnotations(a map[string]interface{}) {
}
