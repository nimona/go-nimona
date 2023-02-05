package nimona

import (
	"fmt"
)

//go:generate ./bin/mockgen -package=nimona -source=codec.go -destination=codec_mock.go

type (
	CodecType string
	Codec     interface {
		// Encode encodes the given value into a byte slice.
		Encode(v DocumentMapper) ([]byte, error)
		// Decode decodes the given byte slice into the given value.
		Decode(b []byte, v DocumentMapper) error
	}
)

const (
	CodecTypeCBOR CodecType = "cbor"
	CodecTypeJSON CodecType = "json"
	CodecTypeYAML CodecType = "yaml"
)

// CodecCBOR is a codec that uses CBOR for encoding and decoding.
type CodecCBOR struct{}

func (c *CodecCBOR) Encode(v DocumentMapper) ([]byte, error) {
	return MarshalCBORBytes(v)
}

func (c *CodecCBOR) Decode(b []byte, v DocumentMapper) error {
	return UnmarshalCBORBytes(b, v)
}

func UnmarshalCBORBytes(b []byte, c DocumentMapper) error {
	m := DocumentMap{}
	err := m.UnmarshalCBOR(b)
	if err != nil {
		return fmt.Errorf("error unmarshaling cbor: %s", err)
	}

	// TODO(geoah): check error
	c.FromDocumentMap(m)
	return nil
}

func MarshalCBORBytes(v DocumentMapper) ([]byte, error) {
	b, err := v.DocumentMap().MarshalCBOR()
	if err != nil {
		return nil, fmt.Errorf("error marshaling cbor: %s", err)
	}
	return b, nil
}

func MustMarshalCBORBytes(v DocumentMapper) []byte {
	b, err := MarshalCBORBytes(v)
	if err != nil {
		panic(err)
	}
	return b
}
