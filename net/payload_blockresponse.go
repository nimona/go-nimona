package net

import (
	blocks "nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&BlockResponse{})
}

// BlockResponse -
type BlockResponse struct {
	RequestID string            `json:"requestID"`
	Block     []byte            `json:"block"`
	Signature *crypto.Signature `json:"-"`
}

func (r *BlockResponse) GetType() string {
	return "blx.response"
}

func (r *BlockResponse) GetSignature() *crypto.Signature {
	return r.Signature
}

func (r *BlockResponse) SetSignature(s *crypto.Signature) {
	r.Signature = s
}

func (r *BlockResponse) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (r *BlockResponse) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
