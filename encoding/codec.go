package encoding

import (
	"errors"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/mr-tron/base58/base58"
	"github.com/ugorji/go/codec"
)

func MarshalBase58(p interface{}) (string, error) {
	b, err := Marshal(p)
	if err != nil {
		return "", err
	}

	return Base58Encode(b), nil
}

// Marshal payload as block
func Marshal(p interface{}) ([]byte, error) {
	block := Encode(p)
	if block == nil {
		return nil, errors.New("could not encode as block")
	}

	spew.Dump(block)

	b := []byte{}
	enc := codec.NewEncoderBytes(&b, CborHandler())
	if err := enc.Encode(block); err != nil {
		return nil, err
	}

	return b, nil
}

// Unmarshal anything from cbor
func Unmarshal(b []byte, p interface{}) error {
	block := &Block{}
	dec := codec.NewDecoderBytes(b, CborHandler())
	if err := dec.Decode(block); err != nil {
		return err
	}

	Decode(block, p)

	return nil
}

// CborHandler for un/marshaling blocks
func CborHandler() *codec.CborHandle {
	ch := &codec.CborHandle{}
	ch.Canonical = true
	ch.MapType = reflect.TypeOf(map[string]interface{}(nil))
	return ch
}

// Base58Encode encodes a byte slice b into a base-58 encoded string.
func Base58Encode(b []byte) (s string) {
	return base58.Encode(b)
}

// Base58Decode decodes a base-58 encoded string into a byte slice b.
func Base58Decode(s string) (b []byte, err error) {
	return base58.Decode(s)
}
