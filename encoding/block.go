package encoding

const tagName = "nimona"

type Key struct {
	Alg string
}

type Signature struct {
	Key *Key   `nimona:"key"`
	Alg string `nimona:"alg"`
}

func (b *Signature) MarshalBlock() ([]byte, error) {
	return Marshal(b)
}

func (b *Signature) UnmarshalBlock(bytes []byte) error {
	return Unmarshal(bytes, b)
}

// Policy for Block
type Policy struct {
	Description string   `nimona:"description" json:"description,omitempty"`
	Subjects    []*Key   `nimona:"subjects" json:"subjects"`
	Actions     []string `nimona:"actions" json:"actions"`
	Effect      string   `nimona:"effect" json:"effect"`
}

// Metadata for Block
type Metadata struct {
	Parent    string   `nimona:"parentID" json:"parentID,omitempty"`
	Ephemeral bool     `nimona:"ephemeral" json:"ephemeral,omitempty"`
	Policies  []Policy `nimona:"policies" json:"policies,omitempty"`
}

// Block for exchanging data via the exchange
type Block struct {
	Type      string            `nimona:"@type" json:"type,omitempty"`
	Headers   map[string]string `nimona:"headers" json:"headers,omitempty"`
	Metadata  *Metadata         `nimona:"metadata" json:"metadata,omitempty"`
	Payload   interface{}       `nimona:"payload" json:"payload,omitempty"`
	Signature []byte            `nimona:"@signature" json:"signature,omitempty"`
}
