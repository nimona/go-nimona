package blocks

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"math/big"

	"github.com/ugorji/go/codec"
)

func ParseUnmarshalOptions(opts ...UnmarshalOption) *UnmarshalOptions {
	options := &UnmarshalOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

type UnmarshalOptions struct {
	Verify      bool
	ReturnBlock bool
}

type UnmarshalOption func(*UnmarshalOptions)

func Verify() UnmarshalOption {
	return func(opts *UnmarshalOptions) {
		opts.Verify = true
	}
}

func ReturnBlock() UnmarshalOption {
	return func(opts *UnmarshalOptions) {
		opts.ReturnBlock = true
	}
}

func Unmarshal(b []byte, opts ...UnmarshalOption) (interface{}, error) {
	options := ParseUnmarshalOptions(opts...)

	tb := &Block{}
	dec := codec.NewDecoderBytes(b, CborHandler())
	if err := dec.Decode(tb); err != nil {
		return nil, err
	}

	signatureBytes := tb.Signature
	tb.Signature = nil

	// verify
	if options.Verify {
		digest, err := getDigest(tb)
		if err != nil {
			return nil, err
		}

		si, err := Unmarshal(signatureBytes)
		if err != nil {
			return nil, err
		}

		signature := si.(*Signature)
		mKey := signature.Key.Materialize()
		pKey, ok := mKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("only ecdsa public keys are currently supported")
		}

		// TODO implement more algorithms
		if signature.Alg != ES256 {
			return nil, ErrAlgorithNotImplemented
		}

		hash := sha256.Sum256(digest)
		rBytes := new(big.Int).SetBytes(signature.Signature[0:32])
		sBytes := new(big.Int).SetBytes(signature.Signature[32:64])

		if ok := ecdsa.Verify(pKey, hash[:], rBytes, sBytes); !ok {
			return nil, ErrCouldNotVerify
		}
	}

	// unmarshal
	o := &Block{
		Type:    tb.Type,
		Payload: map[string]interface{}{},
	}

	dec = codec.NewDecoderBytes(b, CborHandler())
	if err := dec.Decode(o); err != nil {
		return nil, err
	}

	t := GetType(tb.Type)
	v := TypeToPtrInterface(t)

	DecodeInto(o, v)

	o.Payload = v

	if options.ReturnBlock {
		return o, nil
	}

	return v, nil
}

// UnmarshalInto something from cbor
func UnmarshalInto(b []byte, p interface{}) error {
	block := &Block{}
	dec := codec.NewDecoderBytes(b, CborHandler())
	if err := dec.Decode(block); err != nil {
		return err
	}

	DecodeInto(block, p)

	return nil
}
