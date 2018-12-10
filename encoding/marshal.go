package encoding

import (
	"github.com/ugorji/go/codec"
)

// Marshal encodes an object to cbor
func Marshal(o *Object) ([]byte, error) {
	b := []byte{}
	enc := codec.NewEncoderBytes(&b, CborHandler())
	err := enc.Encode(o)
	return b, err
}
