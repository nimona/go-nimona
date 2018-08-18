package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	"github.com/nimona/go-nimona/blocks"
	"github.com/pkg/errors"
)

// Key defines the minimal interface for each of the
// key types.
type Key interface {
	// Materialize creates the corresponding key. For example,
	// RSA types would create *rsa.PublicKey or *rsa.PrivateKey,
	// EC types would create *ecdsa.PublicKey or *ecdsa.PrivateKey,
	// and OctetSeq types create a []byte key.
	Materialize() (interface{}, error)

	// Marshal returns a CBOR encoded blocks.Block with the indicated
	// hashing algorithm, according to JWK (RFC 7638)
	Marshal() ([]byte, error)
}

// New creates a Key from the given key.
func New(key interface{}) (Key, error) {
	if key == nil {
		return nil, errors.New("missing key")
	}

	switch v := key.(type) {
	// case *rsa.PrivateKey:
	// 	return newRSAPrivateKey(v)
	// case *rsa.PublicKey:
	// 	return newRSAPublicKey(v)
	case *ecdsa.PrivateKey:
		return newECDSAPrivateKey(v)
	case *ecdsa.PublicKey:
		return newECDSAPublicKey(v)
	// case []byte:
	// 	return newSymmetricKey(v)
	default:
		return nil, errors.Errorf(`invalid key type %T`, key)
	}
}

// KeyFromBlock returns a Key from a blocks.Block of type key.
func KeyFromBlock(k *blocks.Block) (Key, error) {
	if k.Type != "key" {
		return nil, errors.New("invalid blocks.Block type")
	}

	h := k.Payload.(Headers)
	return KeyFromHeaders(&h)
}

// KeyFromEncodedBlock returns a Key from an ID string.
func KeyFromEncodedBlock(id string) (Key, error) {
	b, err := blocks.Base58Decode(id)
	if err != nil {
		return nil, err
	}

	block := &blocks.Block{}
	if err := blocks.Unmarshal(b, block); err != nil {
		return nil, err
	}

	return KeyFromBlock(block)
}

// KeyFromHeaders returns a Key from a Key Headers.
func KeyFromHeaders(h *Headers) (Key, error) {
	var curve elliptic.Curve
	switch h.Curve {
	case P256:
		curve = elliptic.P256()
	case P384:
		curve = elliptic.P384()
	case P521:
		curve = elliptic.P521()
	default:
		return nil, errors.Errorf(`invalid curve name %s`, h.Curve)
	}

	var key Key
	switch h.KeyType {
	case EC:
		if len(h.D) > 0 {
			key = &ECDSAPrivateKey{
				headers: h,
				key: &ecdsa.PrivateKey{
					PublicKey: ecdsa.PublicKey{
						Curve: curve,
						X:     bigIntFromBytes(h.X),
						Y:     bigIntFromBytes(h.Y),
					},
					D: bigIntFromBytes(h.D),
				},
			}
		} else {
			key = &ECDSAPublicKey{
				headers: h,
				key: &ecdsa.PublicKey{
					Curve: curve,
					X:     bigIntFromBytes(h.X),
					Y:     bigIntFromBytes(h.Y),
				},
			}
		}
	default:
		return nil, errors.Errorf(`invalid kty %s`, h.KeyType)
	}

	return key, nil
}

func bigIntFromBytes(b []byte) *big.Int {
	i := &big.Int{}
	return i.SetBytes(b)
}
