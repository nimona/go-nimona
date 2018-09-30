package net

import (
	"nimona.io/go/primitives"
)

type HandshakeSynAck struct {
	Nonce     string                `json:"nonce"`
	Signature *primitives.Signature `json:"-"`
}

func (r *HandshakeSynAck) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/handshake.syn-ack",
		Payload: map[string]interface{}{
			"nonce": r.Nonce,
		},
		Signature: r.Signature,
	}
}

func (r *HandshakeSynAck) FromBlock(block *primitives.Block) {
	r.Nonce = block.Payload["nonce"].(string)
	r.Signature = block.Signature
}
