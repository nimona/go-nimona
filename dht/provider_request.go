package dht // import "nimona.io/go/dht"

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/primitives"
)

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string                `json:"requestID,omitempty" mapstructure:"requestID,omitempty"`
	Key       string                `json:"key" mapstructure:"key"`
	Signature *primitives.Signature `json:"signature" mapstructure:"-"`
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
	mapstructure.Decode(block.Payload, r)
	r.Signature = block.Signature
}
