package nimona

//go:generate ./bin/mockgen -package=nimona -source=codec.go -destination=codec_mock.go

type (
	// CodecType string
	Codec interface {
		Marshal(v *DocumentMap) ([]byte, error)
		Unmarshal(b []byte, v *DocumentMap) error
	}
)

type CodecJSON struct{}

func (c *CodecJSON) Marshal(v *DocumentMap) ([]byte, error) {
	return v.MarshalJSON()
}

func (c *CodecJSON) Unmarshal(b []byte, v *DocumentMap) error {
	return v.UnmarshalJSON(b)
}
