package keys

import (
	"crypto/ecdsa"
	"errors"

	"github.com/nimona/go-nimona/blocks"
)

// ECDSAPrivateKey is a type of CWK generated from ECDH-ES private keys
type ECDSAPrivateKey struct {
	headers *Headers
	key     *ecdsa.PrivateKey
}

// Materialize returns the EC-DSA private key represented by this JWK
func (k ECDSAPrivateKey) Materialize() interface{} {
	return k.key
}

// Marshal into an encoded blocks.Block
func (k ECDSAPrivateKey) Marshal() (buf []byte, err error) {
	b := &blocks.Block{
		Type:    "key",
		Payload: k.headers,
	}
	return blocks.Marshal(b)
}

func newECDSAPrivateKey(key *ecdsa.PrivateKey) (*ECDSAPrivateKey, error) {
	if key == nil {
		return nil, errors.New(`non-nil ecdsa.PrivateKey required`)
	}
	return &ECDSAPrivateKey{
		headers: &Headers{
			KeyType: EC,
			Curve:   key.Curve.Params().Name,
			X:       key.X.Bytes(),
			Y:       key.Y.Bytes(),
			D:       key.D.Bytes(),
		},
		key: key,
	}, nil
}
