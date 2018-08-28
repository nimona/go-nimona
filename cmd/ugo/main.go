package main

import (
	"github.com/ugorji/go/codec"
)

type A struct {
	A string
	B B
}

func (a *A) MarshalBinary() (text []byte, err error) {
	lbs := []byte{}
	enc := codec.NewEncoderBytes(&lbs, &codec.JsonHandle{})
	enc.Encode([]interface{}{
		a.A,
		a.B,
	})
	return lbs, nil
}

func (a *A) UnmarshalBinary(text []byte) (err error) {
	return nil
}

type B struct{}

func (b *B) MarshalBinary() (text []byte, err error) {
	return []byte("B"), nil
}

func (b *B) UnmarshalBinary(text []byte) (err error) {
	return nil
}

func main() {
	a := A{
		A: "A",
		B: B{},
	}
	lbs := []byte{}
	enc := codec.NewEncoderBytes(&lbs, &codec.JsonHandle{})
	enc.Encode(a)
}
