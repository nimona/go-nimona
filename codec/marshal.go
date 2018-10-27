package codec // import "nimona.io/go/codec"

import (
	"github.com/ugorji/go/codec"
)

// Marshal anything into cbor
func Marshal(o interface{}) ([]byte, error) {
	b := []byte{}
	enc := codec.NewEncoderBytes(&b, CborHandler())
	if err := enc.Encode(o); err != nil {
		return nil, err
	}

	return b, nil
}
