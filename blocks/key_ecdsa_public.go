package blocks

import (
	"crypto/ecdsa"
	"errors"
)

// ECDSAPublicKey is a type of CWK generated from ECDSA public keys
type ECDSAPublicKey struct {
	headers *KeyHeaders
	key     *ecdsa.PublicKey
}

// Materialize returns the EC-DSA public key represented by this JWK
func (k ECDSAPublicKey) Materialize() interface{} {
	return k.key
}

// Marshal into an encoded Block
func (k ECDSAPublicKey) Marshal() (buf []byte, err error) {
	b := &Block{
		Type:    "key",
		Payload: k.headers,
	}
	return Marshal(b)
}

func newECDSAPublicKey(key *ecdsa.PublicKey) (*ECDSAPublicKey, error) {
	if key == nil {
		return nil, errors.New(`non-nil ecdsa.PublicKey required`)
	}

	return &ECDSAPublicKey{
		headers: &KeyHeaders{
			KeyType: EC,
			Curve:   key.Curve.Params().Name,
			X:       key.X.Bytes(),
			Y:       key.Y.Bytes(),
		},
		key: key,
	}, nil
}
