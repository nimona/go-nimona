package net // import "nimona.io/go/net"

import (
	"nimona.io/go/primitives"
)

// BlockRequest payload for BlockRequestType
type BlockRequest struct {
	RequestID string
	ID        string
	Signature *primitives.Signature
	Sender    *primitives.Key
	response  chan interface{}
}

func (r *BlockRequest) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/block.request",
		Payload: map[string]interface{}{
			"requestID": r.RequestID,
			"blockID":   r.ID,
		},
		Signature: r.Signature,
	}
}

func (r *BlockRequest) FromBlock(block *primitives.Block) {
	r.ID = block.Payload["blockID"].(string)
	r.RequestID = block.Payload["requestID"].(string)
	r.Signature = block.Signature
	r.Sender = r.Signature.Key
}
