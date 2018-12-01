package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	"github.com/pkg/errors"
	"nimona.io/go/encoding"
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

//go:generate go run nimona.io/go/cmd/objectify -schema /key -type Key -out key_generated.go

// Key defines the minimal interface for each of the
// key types.
type Key struct {
	Algorithm              string `json:"alg,omitempty"`
	KeyID                  string `json:"kid,omitempty"`
	KeyType                string `json:"kty,omitempty"`
	KeyUsage               string `json:"use,omitempty"`
	KeyOps                 string `json:"key_ops,omitempty"`
	X509CertChain          string `json:"x5c,omitempty"`
	X509CertThumbprint     string `json:"x5t,omitempty"`
	X509CertThumbprintS256 string `json:"x5tS256,omitempty"`
	X509URL                string `json:"x5u,omitempty"`
	Curve                  string `json:"crv,omitempty"`
	X                      []byte `json:"x,omitempty"`
	Y                      []byte `json:"y,omitempty"`
	D                      []byte `json:"d,omitempty"`

	RawObject *encoding.Object `json:"@"`
}

// NewKeyFromObject returns a key from an object
func NewKeyFromObject(o *encoding.Object) (*Key, error) {
	// TODO check type?
	p := &Key{}
	if err := o.Unmarshal(p); err != nil {
		return nil, err
	}

	p.RawObject = o

	return p, nil
}

// Hash of the key
func (k *Key) Hash() []byte {
	return k.ToObject().Hash()
}

// HashBase58 of the key
func (k *Key) HashBase58() string {
	return k.ToObject().HashBase58()
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

	key := &Key{}

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
