package dht

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/primitives"
)

// Provider payload
type Provider struct {
	BlockIDs  []string              `json:"blockIDs" mapstructure:"blockIDs"`
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
	mapstructure.Decode(block.Payload, p)
	p.Signature = block.Signature
}
