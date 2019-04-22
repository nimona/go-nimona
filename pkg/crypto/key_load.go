package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

// GenerateKey creates a new ecdsa private key
func GenerateKey() (*PrivateKey, error) {
	mprv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	prv, err := NewPrivateKey(mprv)
	if err != nil {
		return nil, err
	}

	return prv, nil
}
