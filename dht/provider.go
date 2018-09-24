package dht

import (
	"nimona.io/go/primitives"
)

// Provider payload
type Provider struct {
	BlockIDs  []string              `json:"blockIDs"`
	Signature *primitives.Signature `json:"-"`
}

func (p *Provider) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/dht.provider",
		Payload: map[string]interface{}{
			"blockIDs": p.BlockIDs,
		},
		Signature: p.Signature,
	}
}

func (p *Provider) FromBlock(block *primitives.Block) {
	p.BlockIDs = block.Payload["blockIDs"].([]string)
	p.Signature = block.Signature
}
