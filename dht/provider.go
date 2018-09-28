package dht

import (
	"github.com/mitchellh/mapstructure"
	ucodec "github.com/ugorji/go/codec"

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

// CodecDecodeSelf helper for cbor unmarshaling
func (p *Provider) CodecDecodeSelf(dec *ucodec.Decoder) {
	b := &primitives.Block{}
	dec.MustDecode(b)
	p.FromBlock(b)
}

// CodecEncodeSelf helper for cbor marshaling
func (p *Provider) CodecEncodeSelf(enc *ucodec.Encoder) {
	b := p.Block()
	enc.MustEncode(b)
}
