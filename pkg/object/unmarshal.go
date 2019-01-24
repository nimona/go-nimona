package object

import (
	"github.com/ugorji/go/codec"
)

// Unmarshal a cbor encoded block (container) into a registered type, or map
func Unmarshal(b []byte) (*Object, error) {
	o := New()
	dec := codec.NewDecoderBytes(b, CborHandler())
	if err := dec.Decode(o); err != nil {
		return nil, err
	}

	return o, nil
}
