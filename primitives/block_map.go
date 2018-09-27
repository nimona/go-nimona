package primitives

import (
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

func BlockToMap(block *Block) map[string]interface{} {
	t := struct {
		Type        string                 `json:"type,omitempty" mapstructure:"type,omitempty"`
		Annotations *Annotations           `json:"annotations,omitempty" mapstructure:"annotations,omitempty"`
		Payload     map[string]interface{} `json:"payload,omitempty" mapstructure:"payload,omitempty"`
		Signature   map[string]interface{} `json:"signature,omitempty" mapstructure:"signature,omitempty"`
	}{
		Type:        block.Type,
		Annotations: block.Annotations,
		Payload:     block.Payload,
	}
	if block.Signature != nil {
		t.Signature = BlockToMap(block.Signature.Block())
	}
	s := structs.New(t)
	s.TagName = "mapstructure"
	m := s.Map()
	return m
}

func BlockFromMap(m map[string]interface{}) *Block {
	t := struct {
		Type        string                 `json:"type,omitempty" mapstructure:"type,omitempty"`
		Annotations *Annotations           `json:"annotations,omitempty" mapstructure:"annotations,omitempty"`
		Payload     map[string]interface{} `json:"payload,omitempty" mapstructure:"payload,omitempty"`
		Signature   map[string]interface{} `json:"signature,omitempty" mapstructure:"signature,omitempty"`
	}{}
	if err := mapstructure.Decode(m, &t); err != nil {
		panic(err)
	}
	b := &Block{
		Type:        t.Type,
		Annotations: t.Annotations,
		Payload:     t.Payload,
	}
	if t.Signature != nil {
		b.Signature = &Signature{}
		sigBlock := BlockFromMap(t.Signature)
		b.Signature.FromBlock(sigBlock)
	}
	// TODO(geoah) error
	return b
}
