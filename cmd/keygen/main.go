package main

import (
	"fmt"
	"os"

	"nimona.io/pkg/crypto"
)

func main() {
	var k crypto.PrivateKey
	if len(os.Args) > 1 {
		k = crypto.PrivateKey(os.Args[1])
	} else {
		k, _ = crypto.GenerateEd25519PrivateKey() // nolint: errcheck
	}
	fmt.Println("private key:", k.String())
	fmt.Println("public key:", k.PublicKey().String())
}
