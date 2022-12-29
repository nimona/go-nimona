package nimona

type Request struct {
	Type    string             `cborgen:"$type"`
	Body    []byte             `cborgen:"-"`
	Codec   Codec              `cborgen:"-"`
	Decode  func(Cborer) error `cborgen:"-"`
	Respond func(Cborer) error `cborgen:"-"`
}
