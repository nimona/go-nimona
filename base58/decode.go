package base58

import "github.com/mr-tron/base58/base58"

// Decode decodes a base-58 encoded string into a byte slice b.
func Decode(s string) (b []byte, err error) {
	return base58.Decode(s)
}
