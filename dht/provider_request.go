package dht

import (
	"nimona.io/go/primitives"
)

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string                `json:"requestID,omitempty"`
	Key       string                `json:"key"`
	Signature *primitives.Signature `json:"-"`
}

func (r *ProviderRequest) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/dht.provider.request",
		Payload: map[string]interface{}{
			"requestID": r.RequestID,
			"key":       r.Key,
		},
		Signature: r.Signature,
	}
}

func (r *ProviderRequest) FromBlock(block *primitives.Block) {
	r.RequestID = block.Payload["requestID"].(string)
	r.Key = block.Payload["key"].(string)
	r.Signature = block.Signature
}
