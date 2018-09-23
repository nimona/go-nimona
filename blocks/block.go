package blocks

import "nimona.io/go/crypto"

func init() {
	RegisterContentType(&crypto.Key{})
	RegisterContentType(&crypto.Signature{})
}

const defaultTag = "json"

// Policy for Block
type Policy struct {
	Description string   `json:"description,omitempty"`
	Subjects    []string `json:"subjects"`
	Actions     []string `json:"actions"`
	Effect      string   `json:"effect"`
}

// type Headers struct{}

// Annotations for Block
type Annotations struct {
	Parent   string   `json:"parentID,omitempty"`
	Policies []Policy `json:"policies,omitempty"`
}

// Block for exchanging data via the exchange
type Block struct {
	Type        string                 `json:"type,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Signature   map[string]interface{} `json:"signature,omitempty"`
}

func (block *Block) Map() map[string]interface{} {
	m := map[string]interface{}{}
	if block.Annotations != nil {
		m["annotations"] = block.Annotations
	}
	if block.Payload != nil {
		m["payload"] = block.Payload
	}
	if block.Signature != nil {
		m["signature"] = block.Signature
	}
	if block.Type != "" {
		m["type"] = block.Type
	}
	return m
}

// ID calculated the id for the block
func (block *Block) ID() string {
	d, err := getDigest(block)
	if err != nil {
		panic(err)
	}

	hash, err := SumSha3(d)
	if err != nil {
		panic(err)
	}

	return hash
}

func ID(v Typed) string {
	b, err := Pack(v)
	if err != nil {
		panic(err)
	}

	d, err := getDigest(b)
	if err != nil {
		panic(err)
	}

	hash, err := SumSha3(d)
	if err != nil {
		panic(err)
	}

	return string(hash)
}
