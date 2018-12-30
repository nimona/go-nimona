package encoding

import (
	"github.com/ugorji/go/codec"
)

// UnmarshalSimple unmarshals a cbor encoded object into given interface
func UnmarshalSimple(b []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(b, CborHandler())
	return dec.Decode(v)
}
