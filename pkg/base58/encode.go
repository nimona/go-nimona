package base58

import "github.com/mr-tron/base58/base58"

// Encode encodes a byte slice b into a base-58 encoded string.
func Encode(b []byte) (s string) {
	return base58.Encode(b)
}
