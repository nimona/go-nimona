package net // import "nimona.io/go/net"

import (
	"nimona.io/go/primitives"
)

type HandshakeAck struct {
	Nonce     string                `json:"nonce"`
	Signature *primitives.Signature `json:"-"`
}

func (r *HandshakeAck) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/handshake.ack",
		Payload: map[string]interface{}{
			"nonce": r.Nonce,
		},
		Signature: r.Signature,
	}
}

func (r *HandshakeAck) FromBlock(block *primitives.Block) {
	r.Nonce = block.Payload["nonce"].(string)
	r.Signature = block.Signature
}
