package primitives

import (
	"nimona.io/go/codec"
)

func BlockToMap(b *Block) map[string]interface{} {
	bs, err := codec.Marshal(b)
	if err != nil {
		panic(err)
	}
	m := map[string]interface{}{}
	if err := codec.Unmarshal(bs, &m); err != nil {
		panic(err)
	}
	return m
}

func BlockFromMap(m map[string]interface{}) *Block {
	bs, err := codec.Marshal(m)
	if err != nil {
		panic(err)
	}
	b := &Block{}
	if err := codec.Unmarshal(bs, b); err != nil {
		panic(err)
	}
	return b
}
