package primitives

import (
	"nimona.io/go/codec"
)

func Unmarshal(bytes []byte) (*Block, error) {
	t := &struct {
		Type        string                 `json:"type,omitempty" mapstructure:"type,omitempty"`
		Annotations *Annotations           `json:"annotations,omitempty" mapstructure:"annotations,omitempty"`
		Payload     map[string]interface{} `json:"payload,omitempty" mapstructure:"payload,omitempty"`
		Signature   map[string]interface{} `json:"signature,omitempty" mapstructure:"signature,omitempty"`
	}{}
	if err := codec.Unmarshal(bytes, t); err != nil {
		return nil, err
	}
	b := &Block{
		Type:        t.Type,
		Annotations: t.Annotations,
		Payload:     t.Payload,
	}
	if t.Signature != nil {
		sigBlock := BlockFromMap(t.Signature)
		sig := &Signature{}
		sig.FromBlock(sigBlock)
		b.Signature = sig
	}
	return b, nil
}
