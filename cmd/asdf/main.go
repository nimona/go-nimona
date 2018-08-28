package main

import (
	"fmt"

	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/codec"
)

type Marshalable interface {
	// MarshalBinary() ([]byte, error)
	// UnmarshalBinary([]byte) error
	// MarshalText() ([]byte, error)
	// UnmarshalText([]byte) error
}

type Key struct {
	Key string
}

func (k Key) MarshalBinary() ([]byte, error) {
	return []byte(k.Key), nil
}
func (k Key) UnmarshalBinary([]byte) error {
	return nil
}

type Signature struct {
	Key Marshalable
	Sig string
}

func (s *Signature) MarshalBinary() ([]byte, error) {
	return []byte(s.Sig), nil
}

func (s *Signature) MarshalText() ([]byte, error) {
	return []byte(s.Sig), nil
}

func newSig() *Signature {
	return &Signature{
		Key: &Key{
			Key: "a",
		},
		Sig: "b",
	}
}

func main() {
	s := newSig()
	b, err := codec.Marshal(s)
	if err != nil {
		panic(err)
	}
	fmt.Println(blocks.Base58Encode(b))
}
