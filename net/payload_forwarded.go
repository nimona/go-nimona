package net

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/primitives"
)

// ForwardRequest is the payload for proxied blocks
type ForwardRequest struct {
	Recipient *primitives.Key
	FwBlock   *primitives.Block
	Signature *primitives.Signature
}

func (r *ForwardRequest) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/block.forward.request",
		Payload: map[string]interface{}{
			"recipient": r.Recipient.Block(),
			"block":     r.Block,
		},
		Signature: r.Signature,
	}
}

func (r *ForwardRequest) FromBlock(block *primitives.Block) {
	t := &struct {
		Recipient *primitives.Block `mapstructure:"recipient,omitempty"`
		FwBlock   *primitives.Block `mapstructure:"block,omitempty"`
	}{}

	mapstructure.Decode(block.Payload, t)

	if t.Recipient != nil {
		r.Recipient = &primitives.Key{}
		r.Recipient.FromBlock(t.Recipient)
	}
}
