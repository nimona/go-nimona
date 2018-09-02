package blocks

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	"github.com/pkg/errors"
)

// Supported values for KeyType
const (
	EC             = "EC"  // Elliptic Curve
	InvalidKeyType = ""    // Invalid KeyType
	OctetSeq       = "oct" // Octet sequence (used to represent symmetric keys)
	RSA            = "RSA" // RSA
)

const (
	// P256 curve
	P256 string = "P-256"
	// P384 curve
	P384 string = "P-384"
	// P521 curve
	P521 string = "P-521"
)

func init() {
	RegisterContentType("key", Key{})
}

// Key defines the minimal interface for each of the
// key types.
type Key struct {
	Algorithm              string `nimona:"alg,omitempty" json:"alg,omitempty"`
	KeyID                  string `nimona:"kid,omitempty" json:"kid,omitempty"`
	KeyType                string `nimona:"kty,omitempty" json:"kty,omitempty"`
	KeyUsage               string `nimona:"use,omitempty" json:"use,omitempty"`
	KeyOps                 string `nimona:"key_ops,omitempty" json:"key_ops,omitempty"`
	X509CertChain          string `nimona:"x5c,omitempty" json:"x5c,omitempty"`
	X509CertThumbprint     string `nimona:"x5t,omitempty" json:"x5t,omitempty"`
	X509CertThumbprintS256 string `nimona:"x5tS256,omitempty" json:"x5tS256,omitempty"`
	X509URL                string `nimona:"x5u,omitempty" json:"x5u,omitempty"`
	Curve                  string `nimona:"crv,omitempty" json:"crv,omitempty"`
	X                      []byte `nimona:"x,omitempty" json:"x,omitempty"`
	Y                      []byte `nimona:"y,omitempty" json:"y,omitempty"`
	D                      []byte `nimona:"d,omitempty" json:"d,omitempty"`
	key                    interface{}
}

func (b *Key) MarshalBlock() (string, error) {
	bytes, err := Marshal(b)
	if err != nil {
		return "", err
	}
	return Base58Encode(bytes), nil
}

func (b *Key) UnmarshalBlock(b58bytes string) error {
	bytes, err := Base58Decode(b58bytes)
	if err != nil {
		return err
	}
	return UnmarshalInto(bytes, b)
}

func (k *Key) Thumbprint() string {
	t, err := k.MarshalBlock()
	if err != nil {
		panic(err)
	}
	return t
}

// GetPublicKey returns the public key
func (k *Key) GetPublicKey() *Key {
	if len(k.D) == 0 {
		return k
	}

	pk := k.Materialize().(*ecdsa.PrivateKey).Public().(*ecdsa.PublicKey)
	bpk, err := NewKey(pk)
	if err != nil {
		panic(err)
	}

	return bpk
}

func (k *Key) Materialize() interface{} {
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

	var key interface{}
	switch k.KeyType {
	case EC:
		if len(k.D) > 0 {
			key = &ecdsa.PrivateKey{
				PublicKey: ecdsa.PublicKey{
					Curve: curve,
					X:     bigIntFromBytes(k.X),
					Y:     bigIntFromBytes(k.Y),
				},
				D: bigIntFromBytes(k.D),
			}
		} else {
			key = &ecdsa.PublicKey{
				Curve: curve,
				X:     bigIntFromBytes(k.X),
				Y:     bigIntFromBytes(k.Y),
			}
		}
	default:
		panic("invalid kty")
		// return nil, errors.Errorf(`invalid kty %s`, h.KeyType)
	}

	return key
}

// NewKey creates a Key from the given key.
func NewKey(k interface{}) (*Key, error) {
	if k == nil {
		return nil, errors.New("missing key")
	}

	key := &Key{
		key: k,
	}

	switch v := k.(type) {
	// case *rsa.PrivateKey:
	// 	return newRSAPrivateKey(v)
	// case *rsa.PublicKey:
	// 	return newRSAPublicKey(v)
	case *ecdsa.PrivateKey:
		key.KeyType = EC
		key.Curve = v.Curve.Params().Name
		key.X = v.X.Bytes()
		key.Y = v.Y.Bytes()
		key.D = v.D.Bytes()
	case *ecdsa.PublicKey:
		key.KeyType = EC
		key.Curve = v.Curve.Params().Name
		key.X = v.X.Bytes()
		key.Y = v.Y.Bytes()
	// case []byte:
	// 	return newSymmetricKey(v)
	default:
		return nil, errors.Errorf(`invalid key type %T`, key)
	}

	return key, nil
}

func bigIntFromBytes(b []byte) *big.Int {
	i := &big.Int{}
	return i.SetBytes(b)
}
