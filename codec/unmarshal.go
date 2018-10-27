package codec // import "nimona.io/go/codec"

import (
	"github.com/ugorji/go/codec"
)

// Unmarshal anything from cbor
func Unmarshal(b []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(b, CborHandler())
	if err := dec.Decode(v); err != nil {
		return err
	}

	return nil
}
