package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	"github.com/pkg/errors"

	"nimona.io/internal/encoding/base58"
)

//go:generate $GOBIN/objectify -schema /key.public -type PublicKey -in key_public.go -out key_public_generated.go

type PublicKey struct {
	Algorithm string `json:"alg:s,omitempty"`
	// KeyID                  string `json:"kid,omitempty"`
	KeyType string `json:"kty:s,omitempty"`
	// KeyUsage               string `json:"use,omitempty"`
	// KeyOps                 string `json:"key_ops,omitempty"`
	// X509CertChain          string `json:"x5c,omitempty"`
	// X509CertThumbprint     string `json:"x5t,omitempty"`
	// X509CertThumbprintS256 string `json:"x5tS256,omitempty"`
	// X509URL                string `json:"x5u,omitempty"`
	Curve string `json:"crv:s,omitempty"`
	X     []byte `json:"x:d,omitempty"`
	Y     []byte `json:"y:d,omitempty"`

	Signature *Signature `json:"@signature:o,omitempty"`
}

// Fingerprint of the key
func (k *PublicKey) Fingerprint() Fingerprint {
	fp := &PublicKey{
		Algorithm: k.Algorithm,
		KeyType:   k.KeyType,
		Curve:     k.Curve,
		X:         k.X,
		Y:         k.Y,
	}
	return Fingerprint(base58.Encode(fp.ToObject().Hash()))
}

func (k *PublicKey) Key() interface{} {
	// TODO cache on k.key
	var curve elliptic.Curve
	switch k.Curve {
	case P256:
		curve = elliptic.P256()
	case P384:
		curve = elliptic.P384()
	case P521:
		curve = elliptic.P521()
	default:
		panic("invalid curve name " + k.Curve)
		// return nil, errors.Errorf(`invalid curve name %s`, h.Curve)
	}

	switch k.KeyType {
	case EC:
		return &ecdsa.PublicKey{
			Curve: curve,
			X:     bigIntFromBytes(k.X),
			Y:     bigIntFromBytes(k.Y),
		}
	default:
		panic("invalid kty")
		// return nil, errors.Errorf(`invalid kty %s`, h.KeyType)
	}
}

// NewPublicKey creates a PublicKey from the given key.
func NewPublicKey(key interface{}) (*PublicKey, error) {
	if key == nil {
		return nil, errors.New("missing key")
	}

	k := &PublicKey{}

	switch v := key.(type) {
	// case *rsa.PublicKey:
	// 	return newRSAPublicKey(v)
	case *ecdsa.PublicKey:
		k.KeyType = EC
		k.Curve = v.Curve.Params().Name
		k.X = v.X.Bytes()
		k.Y = v.Y.Bytes()
	default:
		return nil, errors.Errorf(`invalid key type %T`, key)
	}

	return k, nil
}
