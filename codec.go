package nimona

import (
	"bytes"
	"fmt"
)

//go:generate ./bin/mockgen -package=nimona -source=codec.go -destination=codec_mock.go

// TODO(geoah): consider refactoring to use io.Reader and io.Writer
type Codec interface {
	// Encode encodes the given value into a byte slice.
	Encode(v Cborer) ([]byte, error)
	// Decode decodes the given byte slice into the given value.
	Decode(b []byte, v Cborer) error
}

// CodecCBOR is a codec that uses CBOR for encoding and decoding.
type CodecCBOR struct{}

func (c *CodecCBOR) Encode(v Cborer) ([]byte, error) {
	return MarshalCBORBytes(v)
}

func (c *CodecCBOR) Decode(b []byte, v Cborer) error {
	return UnmarshalCBORBytes(b, v)
}

func UnmarshalCBORBytes(b []byte, c Cborer) error {
	err := c.UnmarshalCBOR(bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("error unmarshaling cbor: %s", err)
	}
	return nil
}

func MarshalCBORBytes(v Cborer) ([]byte, error) {
	w := new(bytes.Buffer)
	err := v.MarshalCBOR(w)
	if err != nil {
		return nil, fmt.Errorf("error marshaling cbor: %s", err)
	}
	return w.Bytes(), nil
}

func MustMarshalCBORBytes(v Cborer) []byte {
	b, err := MarshalCBORBytes(v)
	if err != nil {
		panic(err)
	}
	return b
}
