package encoding

import (
	"github.com/ugorji/go/codec"
)

// MarshalSimple marshals anything into a cbor byte stream
func MarshalSimple(v interface{}) ([]byte, error) {
	b := []byte{}
	enc := codec.NewEncoderBytes(&b, CborHandler())
	err := enc.Encode(v)
	return b, err
}
