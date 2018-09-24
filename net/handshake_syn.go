package net

import (
	"nimona.io/go/primitives"
)

type HandshakeSyn struct {
	Nonce     string                `json:"nonce"`
	Signature *primitives.Signature `json:"-"`
}

func (r *HandshakeSyn) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/handshake.syn",
		Payload: map[string]interface{}{
			"nonce": r.Nonce,
		},
		Signature: r.Signature,
	}
}

func (r *HandshakeSyn) FromBlock(block *primitives.Block) {
	r.Nonce = block.Payload["nonce"].(string)
	r.Signature = block.Signature
}
