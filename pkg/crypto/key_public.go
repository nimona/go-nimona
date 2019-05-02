package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	"github.com/pkg/errors"
)

//go:generate go run nimona.io/tools/objectify -schema /key.public -type PublicKey -in key_public.go -out key_public_generated.go

type PublicKey struct {
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

	Signatures []*Signature `json:"sigs,omitempty"`

	Key  interface{} `json:"-"`
	Hash string      `json:"-"`
}

// // Hash of the key
// func (k *PublicKey) Hash() []byte {
// 	return k.ToObject().Hash()
// }

// HashBase58 of the key
func (k *PublicKey) HashBase58() string {
	if k.Hash != "" {
		return k.Hash
	}
	return k.ToObject().HashBase58()
}

// GetPublicKey returns the public key
// func (k *PublicKey) GetPublicKey() *PublicKey {
// 	if len(k.D) == 0 {
// 		return k
// 	}

// 	pk := k.Key.(*ecdsa.PrivateKey).Public().(*ecdsa.PublicKey)
// 	bpk, err := NewKey(pk)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return bpk
// }

func (k *PublicKey) afterFromObject() {
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
		// 	key = &ecdsa.PrivateKey{
		// 		PublicKey: ecdsa.PublicKey{
		// 			Curve: curve,
		// 			X:     bigIntFromBytes(k.X),
		// 			Y:     bigIntFromBytes(k.Y),
		// 		},
		// 		D: bigIntFromBytes(k.D),
		// 	}
		// } else {
		k.Key = &ecdsa.PublicKey{
			Curve: curve,
			X:     bigIntFromBytes(k.X),
			Y:     bigIntFromBytes(k.Y),
		}
		// }
	default:
		panic("invalid kty")
		// return nil, errors.Errorf(`invalid kty %s`, h.KeyType)
	}

	k.Hash = k.ToObject().HashBase58()
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
		k.Key = v
		k.KeyType = EC
		k.Curve = v.Curve.Params().Name
		k.X = v.X.Bytes()
		k.Y = v.Y.Bytes()
	// case []byte:
	// 	return newSymmetricKey(v)
	default:
		return nil, errors.Errorf(`invalid key type %T`, key)
	}

	k.Hash = k.ToObject().HashBase58()

	return k, nil
}
