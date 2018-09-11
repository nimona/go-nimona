package blocks

import (
	"golang.org/x/crypto/sha3"

	"nimona.io/go/base58"
	"nimona.io/go/codec"
)

// SumSha3 returns a base58 encoded SHA3-256 hash of b.
func SumSha3(b []byte) (string, error) {
	d := sha3.Sum256(b)
	y := map[string]interface{}{
		"type":    "sha3.256",
		"payload": d[:],
	}

	h, err := codec.Marshal(y)
	if err != nil {
		return "", err
	}

	return base58.Encode(h[:]), nil
}
