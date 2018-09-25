package primitives

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/base58"
	"nimona.io/go/codec"
)

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
	Annotations *Annotations           `json:"annotations,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Signature   *Signature             `json:"signature,omitempty"`
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

func BlockFromMap(m map[string]interface{}) *Block {
	block := &Block{}
	mapstructure.Decode(m, block)
	// TODO(geoah) error
	return block
}

func BlockFromBase58(s string) (*Block, error) {
	b, err := base58.Decode(s)
	if err != nil {
		return nil, err
	}

	tmpBlock := &struct {
		Type        string                 `json:"type,omitempty"`
		Annotations *Annotations           `json:"annotations,omitempty"`
		Payload     map[string]interface{} `json:"payload,omitempty"`
		Signature   map[string]interface{} `json:"signature,omitempty"`
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
