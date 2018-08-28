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
	Key  *Key
	Sign bool
}

type MarshalOption func(*MarshalOptions)

func SignWith(key *Key) MarshalOption {
	return func(opts *MarshalOptions) {
		opts.Key = key
		opts.Sign = true
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

	if options.Sign && options.Key != nil {
		if err := Sign(block, options.Key); err != nil {
			panic(err)
		}
	}

	// TODO not sure this is needed any more
	// this conversion is needed in order to avoid an endless loop when Encode
	// is being called from the Block's CodecEncodeSelf method, as MarshalBlock
	// also uses the same cbor Marshaler that end up calling CodecEncodeSelf.
	// b := struct {
	// 	Type      string            `nimona:"type,omitempty" json:"type,omitempty"`
	// 	Headers   map[string]string `nimona:"headers,omitempty" json:"headers,omitempty"`
	// 	Metadata  *Metadata         `nimona:"metadata,omitempty" json:"metadata,omitempty"`
	// 	Payload   interface{}       `nimona:"payload,omitempty" json:"payload,omitempty"`
	// 	Signature []byte            `nimona:"signature,omitempty" json:"signature,omitempty"`
	// }{
	// 	Type:      block.Type,
	// 	Headers:   block.Headers,
	// 	Metadata:  block.Metadata,
	// 	Payload:   block.Payload,
	// 	Signature: block.Signature,
	// }
	bytes := []byte{}
	enc := codec.NewEncoderBytes(&bytes, CborHandler())
	if err := enc.Encode(block); err != nil {
		return nil, err
	}

	return bytes, nil
}

// func Unmarshal(b []byte) (interface{}, error) {
// 	tb := &Block{}
// 	dec := codec.NewDecoderBytes(b, CborHandler())
// 	if err := dec.Decode(tb); err != nil {
// 		return nil, err
// 	}

// 	signatureBytes := tb.Signature
// 	tb.Signature = nil

// 	// verify
// 	digest, err := getDigest(tb)
// 	if err != nil {
// 		return nil, err
// 	}

// 	si, err := Unmarshal(signatureBytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	signature := si.(*Signature)
// 	mKey := signature.Key.Materialize()
// 	pKey, ok := mKey.(*ecdsa.PublicKey)
// 	if !ok {
// 		return nil, errors.New("only ecdsa public keys are currently supported")
// 	}

// 	// TODO implement more algorithms
// 	if signature.Alg != ES256 {
// 		return nil, ErrAlgorithNotImplemented
// 	}

// 	hash := sha256.Sum256(digest)
// 	rBytes := new(big.Int).SetBytes(signature.Signature[0:32])
// 	sBytes := new(big.Int).SetBytes(signature.Signature[32:64])

// 	fmt.Println("___________ D", Base58Encode(digest))
// 	fmt.Println("___________ H", Base58Encode(hash[:]))
// 	fmt.Println("___________ R", Base58Encode(signature.Signature[0:32]))
// 	fmt.Println("___________ S", Base58Encode(signature.Signature[32:64]))

// 	if ok := ecdsa.Verify(pKey, hash[:], rBytes, sBytes); !ok {
// 		return nil, ErrCouldNotVerify
// 	}

// 	// unmarshal
// 	o := &Block{
// 		Type:    tb.Type,
// 		Payload: map[string]interface{}{},
// 	}

// 	dec = codec.NewDecoderBytes(b, CborHandler())
// 	if err := dec.Decode(o); err != nil {
// 		return nil, err
// 	}

// 	// spew.Dump(o.Payload)

// 	// var p interface{}
// 	// if cp := GetContentType(tb.Type); cp != nil {
// 	// 	p = cp
// 	// } else {
// 	// 	p = map[string]interface{}{}
// 	// }

// 	t := GetType(tb.Type)
// 	// pt := reflect.PtrTo(t)
// 	// v := reflect.New(pt).Elem().Interface()
// 	// rv := reflect.ValueOf(&v).Elem()
// 	// rvt := rv.Elem().Type().Elem()
// 	// rv.Set(reflect.New(rvt))
// 	v := TypeToPtrInterface(t)

// 	DecodeInto(o, v)

// 	return v, nil
// }

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

// // UnmarshalInto something from cbor
// func UnmarshalInto(b []byte, p interface{}) error {
// 	block := &Block{}
// 	dec := codec.NewDecoderBytes(b, CborHandler())
// 	if err := dec.Decode(block); err != nil {
// 		return err
// 	}

// 	DecodeInto(block, p)

// 	return nil
// }

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
