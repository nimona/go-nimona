package blocks

const tagName = "nimona"

// Policy for Block
type Policy struct {
	Description string   `nimona:"description" json:"description,omitempty"`
	Subjects    []*Key   `nimona:"subjects" json:"subjects"`
	Actions     []string `nimona:"actions" json:"actions"`
	Effect      string   `nimona:"effect" json:"effect"`
}

// type Headers struct{}

// Metadata for Block
type Metadata struct {
	Parent string `nimona:"parentID" json:"parentID,omitempty"`
	// Ephemeral bool     `nimona:"ephemeral" json:"ephemeral,omitempty"`
	Policies []Policy `nimona:"policies" json:"policies,omitempty"`
}

// Block for exchanging data via the exchange
type Block struct {
	Type      string      `nimona:"type,omitempty" json:"type,omitempty"`
	Metadata  *Metadata   `nimona:"metadata,omitempty" json:"metadata,omitempty"`
	Payload   interface{} `nimona:"payload,omitempty" json:"payload,omitempty"`
	Signature string      `nimona:"signature,omitempty" json:"signature,omitempty"`
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

// ID calculated the id for the block
func (block *Block) ID() string {
	return ID(block.Payload)
}

func ID(payload interface{}) string {
	bytes, err := Marshal(payload) //, SkipHeaders())
	if err != nil {
		panic(err)
	}

	hash, err := SumSha3(bytes)
	if err != nil {
		panic(err)
	}

	return string(hash)
}
