package primitives

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type BlockExt struct{}

func (x BlockExt) WriteExt(interface{}) []byte { panic("unsupported") }

func (x BlockExt) ReadExt(interface{}, []byte) { panic("unsupported") }

func (x BlockExt) ConvertExt(v interface{}) interface{} {
	switch b := v.(type) {
	case Block:
		m := map[string]interface{}{
			"type": b.Type,
		}
		if b.Annotations != nil {
			m["annotations"] = b.Annotations
		}
		if b.Payload != nil {
			m["payload"] = b.Payload
		}
		if b.Signature != nil {
			m["signature"] = b.Signature
		}
		return m
	case *Block:
		m := map[string]interface{}{
			"type": b.Type,
		}
		if b.Annotations != nil {
			m["annotations"] = b.Annotations
		}
		if b.Payload != nil {
			m["payload"] = b.Payload
		}
		if b.Signature != nil {
			m["signature"] = b.Signature
		}
		return m
	default:
		panic(fmt.Sprintf("unsupported format for time conversion: expecting time.Time; got %T", v))
	}
}
func (x BlockExt) UpdateExt(dest interface{}, v interface{}) {
	tt := dest.(*Block)
	switch v2 := v.(type) {
	case map[string]interface{}:
		tt.Type = v2["type"].(string)
		tt.Payload = v2["payload"].(map[string]interface{})
		if annMap, ok := v2["annotations"]; ok {
			tt.Annotations = &Annotations{}
			if err := mapstructure.Decode(annMap, tt.Annotations); err != nil {
				panic(err)
			}
		}
		if sigMap, ok := v2["signature"].(map[string]interface{}); ok {
			tt.Signature = &Signature{}
			sigBlock := BlockFromMap(sigMap)
			tt.Signature.FromBlock(sigBlock)
		}
	default:
		panic(fmt.Sprintf("unsupported format for time conversion: expecting int64/uint64; got %T", v))
	}
}
