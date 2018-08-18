package keys

import (
	"crypto/ecdsa"
	"errors"

	"github.com/nimona/go-nimona/blocks"
)

// ECDSAPublicKey is a type of CWK generated from ECDSA public keys
type ECDSAPublicKey struct {
	headers *Headers
	key     *ecdsa.PublicKey
}

// Materialize returns the EC-DSA public key represented by this JWK
func (k ECDSAPublicKey) Materialize() (interface{}, error) {
	return k.key, nil
}

// Marshal into an encoded blocks.Block
func (k ECDSAPublicKey) Marshal() (buf []byte, err error) {
	b := &blocks.Block{
		Type:    "key",
		Payload: k.headers,
	}
	return blocks.Marshal(b)
}

func newECDSAPublicKey(key *ecdsa.PublicKey) (*ECDSAPublicKey, error) {
	if key == nil {
		return nil, errors.New(`non-nil ecdsa.PublicKey required`)
	}

	return &ECDSAPublicKey{
		headers: &Headers{
			KeyType: EC,
			Curve:   key.Curve.Params().Name,
			X:       key.X.Bytes(),
			Y:       key.Y.Bytes(),
		},
		key: key,
	}, nil
}
