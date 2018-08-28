package blocks

const tagName = "nimona"

// Policy for Block
type Policy struct {
	Description string   `nimona:"description" json:"description,omitempty"`
	Subjects    []*Key   `nimona:"subjects" json:"subjects"`
	Actions     []string `nimona:"actions" json:"actions"`
	Effect      string   `nimona:"effect" json:"effect"`
}

type Headers struct{}

// Metadata for Block
type Metadata struct {
	Parent string `nimona:"parentID" json:"parentID,omitempty"`
	// Ephemeral bool     `nimona:"ephemeral" json:"ephemeral,omitempty"`
	Policies []Policy `nimona:"policies" json:"policies,omitempty"`
}

// Block for exchanging data via the exchange
type Block struct {
	Type      string            `nimona:"type,omitempty" json:"type,omitempty"`
	Headers   map[string]string `nimona:"headers,omitempty" json:"headers,omitempty"`
	Metadata  *Metadata         `nimona:"metadata,omitempty" json:"metadata,omitempty"`
	Payload   interface{}       `nimona:"payload,omitempty" json:"payload,omitempty"`
	Signature []byte            `nimona:"signature,omitempty" json:"signature,omitempty"`
}

// NewBlock is a helper function for creating Blocks.
func NewBlock(contentType string, payload interface{}, recipients ...*Key) *Block {
	// TODO do we need to add the owner on the policy as well?
	block := &Block{
		Type:    contentType,
		Payload: payload,
	}
	if len(recipients) > 0 {
		block.Metadata.Policies = []Policy{
			Policy{
				Description: "policy for recipients",
				Subjects:    recipients,
				Actions:     []string{"read"},
				Effect:      "allow",
			},
		}
	}
	return block
}

// // SetHeader pair in block
// func (block *Block) SetHeader(k, v string) {
// 	if block.Headers == nil {
// 		block.Headers = map[string]string{}
// 	}
// 	block.Headers[k] = v
// }

// // GetHeader by key
// func (block *Block) GetHeader(k string) string {
// 	if block.Headers == nil {
// 		return ""
// 	}
// 	return block.Headers[k]
// }

// // IsSigned checks if the block is signed
// func (block *Block) IsSigned() bool {
// 	// TODO make this part of the block and digest?
// 	return block.Signature != nil
// }

// ID calculated the id for the block
func (block *Block) ID() string {
	b, _ := Marshal(block)
	hash, _ := SumSha3(b)
	return hash
}
