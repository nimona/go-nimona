package nimona

import (
	"github.com/fxamacker/cbor/v2"
	_ "github.com/golang/mock/mockgen/model"
)

//go:generate ./bin/mockgen -package=nimona -source=codec.go -destination=codec_mock.go

type Codec interface {
	// Encode encodes the given value into a byte slice.
	Encode(v interface{}) ([]byte, error)
	// Decode decodes the given byte slice into the given value.
	Decode(b []byte, v interface{}) error
}

// CodecCBOR is a codec that uses CBOR for encoding and decoding.
type CodecCBOR struct{}

func (c *CodecCBOR) Encode(v interface{}) ([]byte, error) {
	return cbor.Marshal(v)
}

func (c *CodecCBOR) Decode(b []byte, v interface{}) error {
	return cbor.Unmarshal(b, v)
}
