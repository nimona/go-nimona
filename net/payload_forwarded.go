package net

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&ForwardRequest{})
}

// ForwardRequest is the payload for proxied blocks
type ForwardRequest struct {
	Recipient *crypto.Key       `json:"recipient"`
	Typed     blocks.Typed      `json:"data"`
	Signature *crypto.Signature `json:"-"`
}

func (r *ForwardRequest) GetType() string {
	return "nimona.forwarded"
}

func (r *ForwardRequest) GetSignature() *crypto.Signature {
	return r.Signature
}

func (r *ForwardRequest) SetSignature(s *crypto.Signature) {
	r.Signature = s
}

func (r *ForwardRequest) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (r *ForwardRequest) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
