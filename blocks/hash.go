package blocks

import (
	"golang.org/x/crypto/sha3"
)

// SumSha3 returns a base58 encoded SHA3-256 hash of b.
func SumSha3(b []byte) (string, error) {
	d := sha3.Sum256(b)
	y := &Block{
		Type:    "sha3.256",
		Payload: d[:],
	}

	h, err := Marshal(y)
	if err != nil {
		return "", err
	}

	return Base58Encode(h[:]), nil
}
