package primitives

import (
	"nimona.io/go/codec"
)

func Marshal(block *Block) ([]byte, error) {
	return codec.Marshal(BlockToMap(block))
}
