package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/teserakt-io/golang-ed25519/extra25519"
	"golang.org/x/crypto/curve25519"
)

func generateEd25519PrivateKey() (ed25519.PrivateKey, error) {
	_, k, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func publicEd25519KeyToCurve25519(edPubKey ed25519.PublicKey) []byte {
	var edPk [ed25519.PublicKeySize]byte
	var curveKey [32]byte
	copy(edPk[:], edPubKey)
	if !extra25519.PublicKeyToCurve25519(&curveKey, &edPk) {
		panic("could not convert ed25519 public key to curve25519")
	}

	return curveKey[:]
}

func privateEd25519KeyToCurve25519(edPrivKey ed25519.PrivateKey) []byte {
	var edSk [ed25519.PrivateKeySize]byte
	var curveKey [32]byte
	copy(edSk[:], edPrivKey)
	extra25519.PrivateKeyToCurve25519(&curveKey, &edSk)

	return curveKey[:]
}

func main() {
	a, _ := generateEd25519PrivateKey()
	b, _ := generateEd25519PrivateKey()

	A := a.Public().(ed25519.PublicKey)
	B := b.Public().(ed25519.PublicKey)

	fmt.Printf("Alice private key (a):\t%x\n", a.Seed())
	fmt.Printf("Alice public key (A):\t%x\n", A)

	fmt.Printf("\nBob private key (b):\t%x\n", b.Seed())
	fmt.Printf("Bob public key (B):\t%x\n", B)

	ca := privateEd25519KeyToCurve25519(a)
	cb := privateEd25519KeyToCurve25519(b)

	fmt.Printf("\nAlice curve25519 private key (ca):\t%x\n", a.Seed())
	fmt.Printf("Alice curve25519 public key point (x co-ord):\t%x\n", A)

	cA := publicEd25519KeyToCurve25519(A)
	cB := publicEd25519KeyToCurve25519(B)

	fmt.Printf("\nBob curve25519 private key (cb):\t%x\n", b.Seed())
	fmt.Printf("Bob curve25519 public key point (x co-ord):\t%x\n", B)

	sharedKeyAB, err := curve25519.X25519(ca, cB)
	if err != nil {
		panic("could perform scalar multiplication")
	}
	sharedKeyBA, err := curve25519.X25519(cb, cA)
	if err != nil {
		panic("could perform scalar multiplication")
	}

	fmt.Printf("\nShared key (Alice):\t%x\n", sharedKeyAB)
	fmt.Printf("Shared key (Bob):\t%x\n", sharedKeyBA)
}
