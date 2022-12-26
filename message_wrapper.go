package nimona

import (
	"io"
)

type Cborer interface {
	MarshalCBOR(io.Writer) error
	UnmarshalCBOR(io.Reader) error
}

type MessageWrapper struct {
	Type string `cborgen:"$type"`
}
