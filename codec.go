package nimona

//go:generate ./bin/mockgen -package=nimona -source=codec.go -destination=codec_mock.go

type (
	// CodecType string
	Codec interface {
		Marshal(v *Document) ([]byte, error)
		Unmarshal(b []byte, v *Document) error
	}
)

type CodecJSON struct{}

func (c *CodecJSON) Marshal(v *Document) ([]byte, error) {
	return v.MarshalJSON()
}

func (c *CodecJSON) Unmarshal(b []byte, v *Document) error {
	return v.UnmarshalJSON(b)
}
