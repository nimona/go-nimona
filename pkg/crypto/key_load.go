package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

// GenerateKey creates a new ecdsa private key
func GenerateKey() (*Key, error) {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	key, err := NewKey(pk)
	if err != nil {
		return nil, err
	}

	return key, nil
}
