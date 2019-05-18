package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	"github.com/pkg/errors"
)

//go:generate $GOBIN/objectify -schema /key.private -type PrivateKey -in key_private.go -out key_private_generated.go

// PrivateKey
type PrivateKey struct {
	Algorithm string `json:"alg,omitempty"`
	// KeyID                  string `json:"kid,omitempty"`
	KeyType string `json:"kty,omitempty"`
	// KeyUsage               string `json:"use,omitempty"`
	// KeyOps                 string `json:"key_ops,omitempty"`
	// X509CertChain          string `json:"x5c,omitempty"`
	// X509CertThumbprint     string `json:"x5t,omitempty"`
	// X509CertThumbprintS256 string `json:"x5tS256,omitempty"`
	// X509URL                string `json:"x5u,omitempty"`
	Curve string `json:"crv,omitempty"`
	X     []byte `json:"x,omitempty"`
	Y     []byte `json:"y,omitempty"`
	D     []byte `json:"d,omitempty"`

	PublicKey *PublicKey `json:"pub,omitempty"`
}

// Fingerprint of the key
func (k *PrivateKey) Fingerprint() string {
	return k.PublicKey.Fingerprint()
}

func (k *PrivateKey) Key() interface{} {
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
		// if len(k.D) > 0 {
		sk := &ecdsa.PrivateKey{
			PublicKey: ecdsa.PublicKey{
				Curve: curve,
				X:     bigIntFromBytes(k.X),
				Y:     bigIntFromBytes(k.Y),
			},
			D: bigIntFromBytes(k.D),
		}
		return sk
	default:
		panic("invalid kty")
		// return nil, errors.Errorf(`invalid kty %s`, h.KeyType)
	}
}

// NewPrivateKey creates a PrivateKey from the given key.
func NewPrivateKey(key interface{}) (*PrivateKey, error) {
	if key == nil {
		return nil, errors.New("missing key")
	}

	k := &PrivateKey{}
	switch v := key.(type) {
	// case *rsa.PrivateKey:
	// 	return newRSAPrivateKey(v)
	case *ecdsa.PrivateKey:
		k.KeyType = EC
		k.Curve = v.Curve.Params().Name
		k.X = v.X.Bytes()
		k.Y = v.Y.Bytes()
		k.D = v.D.Bytes()
		pk, err := NewPublicKey(&v.PublicKey)
		if err != nil {
			return nil, err
		}
		k.PublicKey = pk
	// case []byte:
	// 	return newSymmetricKey(v)
	default:
		return nil, errors.Errorf(`invalid key type %T`, key)
	}

	return k, nil
}
