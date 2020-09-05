package main

import (
	"fmt"

	"nimona.io/pkg/crypto"
)

func main() {
	k, _ := crypto.GenerateEd25519PrivateKey() // nolint: errcheck
	fmt.Println("private key:", k.String())
	fmt.Println("public key:", k.PublicKey().String())
}
