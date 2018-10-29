package primitives

import (
	"nimona.io/go/base58"
	"nimona.io/go/codec"
)

const defaultTag = "json"

// Policy for Block
type Policy struct {
	Description string   `json:"description,omitempty" structs:"description,omitempty" mapstructure:"description,omitempty"`
	Subjects    []string `json:"subjects" structs:"subjects" mapstructure:"subjects"`
	Actions     []string `json:"actions" structs:"actions" mapstructure:"actions"`
	Effect      string   `json:"effect" structs:"effect" mapstructure:"effect"`
}

// type Headers struct{}

// Annotations for Block
type Annotations struct {
	Parent   string   `json:"parentID,omitempty" structs:"parentID,omitempty" mapstructure:"parentID,omitempty"`
	Policies []Policy `json:"policies,omitempty" structs:"policies,omitempty" mapstructure:"policies,omitempty"`
}

// Block for exchanging data via the exchange
type Block struct {
	Type        string                 `json:"type,omitempty" mapstructure:"type,omitempty"`
	Annotations *Annotations           `json:"annotations,omitempty" mapstructure:"annotations,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty" mapstructure:"payload,omitempty"`
	Mandate     *Mandate               `json:"mandate,omitempty" mapstructure:"mandate,omitempty"`
	Signature   *Signature             `json:"signature,omitempty" mapstructure:"signature,omitempty"`
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

func ID(v *Block) string {
	d, err := getDigest(v)
	if err != nil {
		panic(err)
	}

	hash, err := SumSha3(d)
	if err != nil {
		panic(err)
	}

	return string(hash)
}

func (block *Block) Digest() ([]byte, error) {
	return getDigest(block)
}

func Digest(v *Block) ([]byte, error) {
	return getDigest(v)
}

func BlockFromBase58(s string) (*Block, error) {
	b, err := base58.Decode(s)
	if err != nil {
		return nil, err
	}

	tmpBlock := &struct {
		Type        string                 `json:"type,omitempty" mapstructure:"type,omitempty"`
		Annotations *Annotations           `json:"annotations,omitempty" mapstructure:"annotations,omitempty"`
		Payload     map[string]interface{} `json:"payload,omitempty" mapstructure:"payload,omitempty"`
		Signature   map[string]interface{} `json:"signature,omitempty" mapstructure:"signature,omitempty"`
	}{}

	if err := codec.Unmarshal(b, tmpBlock); err != nil {
		return nil, err
	}

	block := &Block{
		Type:        tmpBlock.Type,
		Annotations: tmpBlock.Annotations,
		Payload:     tmpBlock.Payload,
	}

	if tmpBlock.Signature != nil {
		block.Signature = &Signature{}
		sigBlock := BlockFromMap(tmpBlock.Signature)
		block.Signature.FromBlock(sigBlock)
	}

	return block, nil
}
