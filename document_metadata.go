package nimona

import (
	cbg "github.com/whyrusleeping/cbor-gen"
)

type (
	Metadata struct {
		Owner     string       `cborgen:"owner,omitempty"`
		Timestamp cbg.CborTime `cborgen:"timestamp,omitempty"`
		Signature Signature    `cborgen:"_signature,omitempty"`
	}
	Signature struct {
		Signer PeerID `cborgen:"signer"`
		X      []byte `cborgen:"x"`
	}
)
