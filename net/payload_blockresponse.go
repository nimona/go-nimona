package net // import "nimona.io/go/net"

import (
	"nimona.io/go/primitives"
)

// BlockResponse -
type BlockResponse struct {
	RequestID      string
	RequestedBlock *primitives.Block
	Signature      *primitives.Signature
	Sender         *primitives.Key
}

func (r *BlockResponse) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/block.response",
		Payload: map[string]interface{}{
			"requestID": r.RequestID,
			"block":     r.RequestedBlock,
		},
		Signature: r.Signature,
	}
}

func (r *BlockResponse) FromBlock(block *primitives.Block) {
	r.RequestID = block.Payload["requestID"].(string)
	r.RequestedBlock = block.Payload["block"].(*primitives.Block)
	r.Signature = block.Signature
}
