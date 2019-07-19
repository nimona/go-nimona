package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"nimona.io/internal/errors"
	"nimona.io/pkg/object"
)

var (
	// ErrCouldNotVerify is returned when the signature doesn't matches the
	// given key
	ErrCouldNotVerify = errors.New("could not verify signature")
)

// Verify object
func Verify(o object.Object) error {
	if o == nil {
		return errors.New("missing object")
	}

	so := o.GetSignature()
	if so == nil {
		return errors.New("missing signature")
	}

	sig, err := GetObjectSignature(o)
	if err != nil {
		return errors.Wrap(
			errors.New("could not get signature"),
			err,
		)
	}

	hash, err := object.ObjectHash(o)
	if err != nil {
		return err
	}

	return verify(sig, hash)
}

// verify a signature given a hash
func verify(sig *Signature, hash []byte) error {
	switch k := sig.PublicKey.Key().(type) {
	case *ecdsa.PublicKey:
		r := new(big.Int).SetBytes(sig.R)
		s := new(big.Int).SetBytes(sig.S)
		if ok := ecdsa.Verify(k, hash, r, s); !ok {
			return ErrCouldNotVerify
		}

	case *ecdsa.PrivateKey:
		r := new(big.Int).SetBytes(sig.R)
		s := new(big.Int).SetBytes(sig.S)
		pk := k.Public().(*ecdsa.PublicKey)
		if ok := ecdsa.Verify(pk, hash, r, s); !ok {
			return ErrCouldNotVerify
		}

	default:
		return fmt.Errorf("verify does not support %T keys", k)
	}

	return nil
}
