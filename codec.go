package nimona

//go:generate ./bin/mockgen -package=nimona -source=codec.go -destination=codec_mock.go

type Codec interface {
	// Encode encodes the given value into a byte slice.
	Encode(v Cborer) ([]byte, error)
	// Decode decodes the given byte slice into the given value.
	Decode(b []byte, v Cborer) error
}

// CodecCBOR is a codec that uses CBOR for encoding and decoding.
type CodecCBOR struct{}

func (c *CodecCBOR) Encode(v Cborer) ([]byte, error) {
	return v.MarshalCBORBytes()
}

func (c *CodecCBOR) Decode(b []byte, v Cborer) error {
	return v.UnmarshalCBORBytes(b)
}
