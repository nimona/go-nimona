package blocks

import (
	"errors"
	"reflect"

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

func ParseMarshalOptions(opts ...MarshalOption) *MarshalOptions {
	options := &MarshalOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

type MarshalOptions struct {
	Key         *Key
	Sign        bool
	SkipHeaders bool
}

type MarshalOption func(*MarshalOptions)

func SignWith(key *Key) MarshalOption {
	return func(opts *MarshalOptions) {
		opts.Key = key
		opts.Sign = true
	}
}
func SkipHeaders() MarshalOption {
	return func(opts *MarshalOptions) {
		opts.SkipHeaders = true
	}
}

// Marshal payload as block
func Marshal(p interface{}, opts ...MarshalOption) ([]byte, error) {
	if p == nil {
		return nil, nil
	}

	block := Encode(p)
	if block == nil {
		return nil, errors.New("could not encode as block")
	}

	// spew.Dump(block)
	return MarshalBlock(block, opts...)
}

func MarshalBlock(block *Block, opts ...MarshalOption) ([]byte, error) {
	options := ParseMarshalOptions(opts...)

	if options.SkipHeaders {
		block.Headers = nil
	}

	if options.Sign && options.Key != nil {
		if err := Sign(block, options.Key); err != nil {
			panic(err)
		}
	}

	bytes := []byte{}
	enc := codec.NewEncoderBytes(&bytes, CborHandler())
	if err := enc.Encode(block); err != nil {
		return nil, err
	}

	return bytes, nil
}

// TODO make both encode and decode accept both ptrs and structs

// used for encoding as all registered types should be strcts
func TypeToPtrInterface(t reflect.Type) interface{} {
	pt := reflect.PtrTo(t)
	v := reflect.New(pt).Elem().Interface()
	rv := reflect.ValueOf(&v).Elem()
	rvt := rv.Elem().Type().Elem()
	rv.Set(reflect.New(rvt))
	return v
}

// used for decoding as all payloads are ptrs already
func TypeToInterface(t reflect.Type) interface{} {
	// pt := reflect.PtrTo(t)
	v := reflect.New(t).Elem().Interface()
	rv := reflect.ValueOf(&v).Elem()
	rvt := rv.Elem().Type().Elem()
	rv.Set(reflect.New(rvt))
	return v
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
