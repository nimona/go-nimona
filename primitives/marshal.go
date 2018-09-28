package primitives

import (
	ucodec "github.com/ugorji/go/codec"
)

func Marshal(o *Block) ([]byte, error) {
	b := []byte{}
	enc := ucodec.NewEncoderBytes(&b, CborHandler())
	if err := enc.Encode(o); err != nil {
		return nil, err
	}

	return b, nil
}
