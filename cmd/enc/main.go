package main

import (
	"fmt"

	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/encoding"
)

func main() {
	s := &encoding.Signature{
		Key: &encoding.Key{
			Alg: "key-alg",
		},
		Alg: "sig-alg",
	}

	b, err := encoding.Marshal(s)
	if err != nil {
		panic(err)
	}

	fmt.Println(blocks.Base58Encode(b))

	sn := &encoding.Signature{}
	if err := encoding.Unmarshal(b, sn); err != nil {
		panic(err)
	}

	fmt.Println(sn.Alg)
	fmt.Println(sn.Key)
}
