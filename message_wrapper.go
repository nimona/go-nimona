package nimona

type Cborer interface {
	// MarshalCBOR(io.Writer) error
	MarshalCBORBytes() ([]byte, error)
	// UnmarshalCBOR(io.Reader) error
	UnmarshalCBORBytes([]byte) error
}

type MessageWrapper struct {
	Type string `cborgen:"$type"`
}
