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
	blockBytes, _ := primitives.Marshal(r.FwBlock)
	return &primitives.Block{
		Type: "nimona.io/block.forward.request",
		Payload: map[string]interface{}{
			"recipient":  r.Recipient,
			"blockBytes": blockBytes,
		},
		Signature: r.Signature,
	}
}

func (r *ForwardRequest) FromBlock(block *primitives.Block) {
	t := &struct {
		Recipient  string `mapstructure:"recipient,omitempty"`
		BlockBytes []byte `mapstructure:"blockBytes,omitempty"`
	}{}

	mapstructure.Decode(block.Payload, t)

	fwBlock, _ := primitives.Unmarshal(t.BlockBytes)

	r.Recipient = t.Recipient
	r.FwBlock = fwBlock
	r.Signature = block.Signature
}
