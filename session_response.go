package nimona

type Response struct {
	Type   string             `cborgen:"$type"`
	Body   []byte             `cborgen:"-"`
	Codec  Codec              `cborgen:"-"`
	Decode func(Cborer) error `cborgen:"-"`
}
