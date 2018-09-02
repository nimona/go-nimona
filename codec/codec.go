package codec

import (
	"reflect"

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

// Unmarshal anything from cbor
func Unmarshal(b []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(b, CborHandler())
	if err := dec.Decode(v); err != nil {
		return err
	}

	return nil
}

// CborHandler for un/marshaling blocks
func CborHandler() *codec.CborHandle {
	ch := &codec.CborHandle{}
	ch.Canonical = true
	ch.MapType = reflect.TypeOf(map[string]interface{}(nil))
	return ch
}
