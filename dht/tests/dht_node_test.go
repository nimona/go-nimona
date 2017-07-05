package dht

import (
	"fmt"
	"testing"

	dht "github.com/nimona/go-nimona-kad-dht"
)

func TestXor(t *testing.T) {
	b1 := []byte("aasasdfdfa")
	b2 := []byte("basdfa")
	// fmt.Printf("b1: %b, \nb2: %b\n---\n", b1, b2)
	res := dht.Xor(b1, b2)
	// fmt.Printf("b1: %b, \nb2: %b\n---\n", b1, b2)
	fmt.Printf("xor: %08b\n", res)
	for _, b := range res {
		fmt.Println(b)
	}
}
