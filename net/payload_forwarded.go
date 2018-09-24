package net

import (
	"nimona.io/go/base58"
	"nimona.io/go/codec"
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
		Annotations: &primitives.Annotations{
			Policies: []primitives.Policy{
				primitives.Policy{
					Subjects: []string{
						r.Recipient.Thumbprint(),
					},
					Actions: []string{"read"},
					Effect:  "allow",
				},
			},
		},
		Signature: r.Signature,
	}
}

func (r *ForwardRequest) FromBlock(block *primitives.Block) {
	// TODO(geoah) this won't work
	r.FwBlock = block.Payload["block"].(*primitives.Block)
	r.Signature = block.Signature

	key := &primitives.Key{}
	subject := block.Annotations.Policies[0].Subjects[0]
	subjectBytes, err := base58.Decode(subject)
	if err != nil {
		return
	}
	codec.Unmarshal(subjectBytes, key)
	r.Recipient = key
}
