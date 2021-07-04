package main

import (
	"fmt"
	"os"

	"nimona.io/pkg/crypto"
)

func main() {
	k := crypto.PrivateKey{}
	if len(os.Args) > 1 {
		if err := k.UnmarshalString(os.Args[1]); err != nil {
			panic(err)
		}
	} else {
		k, _ = crypto.NewEd25519PrivateKey() // nolint: errcheck
	}
	fmt.Println("private key:", k.String())
	fmt.Println("public key:", k.PublicKey().String())
}
