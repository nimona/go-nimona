package net

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/primitives"
)

// ForwardRequest is the payload for proxied blocks
type ForwardRequest struct {
	Recipient string // address
	FwBlock   *primitives.Block
	Signature *primitives.Signature
}

func (r *ForwardRequest) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/block.forward.request",
		Payload: map[string]interface{}{
			"recipient": r.Recipient,
			"block":     r.Block,
		},
		Signature: r.Signature,
	}
}

func (r *ForwardRequest) FromBlock(block *primitives.Block) {
	t := &struct {
		Recipient string            `mapstructure:"recipient,omitempty"`
		FwBlock   *primitives.Block `mapstructure:"block,omitempty"`
	}{}

	mapstructure.Decode(block.Payload, t)
}
