package keystore

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/tilde"
)

type KeyStore interface {
	PutKey(crypto.PrivateKey) error
	GetKey(tilde.Digest) (*crypto.PrivateKey, error)
}
