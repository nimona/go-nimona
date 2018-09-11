package net

import (
	blocks "nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&BlockRequest{})
}

// BlockRequest payload for BlockRequestType
type BlockRequest struct {
	RequestID string            `json:"requestID"`
	ID        string            `json:"id"`
	Signature *crypto.Signature `json:"-"`
	response  chan interface{}
}

func (r *BlockRequest) GetType() string {
	return "blx.request-block"
}

func (r *BlockRequest) GetSignature() *crypto.Signature {
	return r.Signature
}

func (r *BlockRequest) SetSignature(s *crypto.Signature) {
	r.Signature = s
}

func (r *BlockRequest) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (r *BlockRequest) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
