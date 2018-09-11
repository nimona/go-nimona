package blocks

import "nimona.io/go/crypto"

func ParsePackOptions(opts ...PackOption) *PackOptions {
	options := &PackOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

type PackOptions struct {
	SkipTopLevelEncode bool
	EncodeNested       bool
	EncodeNestedBase58 bool
	Key                *crypto.Key
	Sign               bool
}

type PackOption func(*PackOptions)

func EncodeNestedSkipTopLevel() PackOption {
	return func(opts *PackOptions) {
		opts.SkipTopLevelEncode = true
	}
}

func EncodeNested() PackOption {
	return func(opts *PackOptions) {
		opts.EncodeNested = true
	}
}

func EncodeNestedBase58() PackOption {
	return func(opts *PackOptions) {
		opts.EncodeNestedBase58 = true
	}
}

func SignWith(key *crypto.Key) PackOption {
	return func(opts *PackOptions) {
		opts.Key = key
		opts.Sign = true
	}
}
