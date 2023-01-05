package nimona

import (
	cbg "github.com/whyrusleeping/cbor-gen"
)

type (
	Metadata struct {
		Owner     string       `cborgen:"owner"`
		Timestamp cbg.CborTime `cborgen:"timestamp"`
		Signature Signature    `cborgen:"_signature"`
	}
	Signature struct {
		X []byte `cborgen:"x"`
	}
)
