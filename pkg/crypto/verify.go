package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/pkg/errors"

	"nimona.io/go/encoding"
)

var (
	// ErrCouldNotVerify is returned when the signature doesn't matches the
	// given key
	ErrCouldNotVerify = errors.New("could not verify signature")
)

// Verify object given the signer's key
func Verify(o *encoding.Object) error {
	if o == nil {
		return errors.New("missing object")
	}

	so := o.GetSignature()
	if so == nil {
		return errors.New("missing signature")
	}

	ko := o.GetSignerKey()
	if so == nil {
		return errors.New("missing signer key")
	}

	sig := &Signature{}
	if err := sig.FromObject(so); err != nil {
		return err
	}

	key := &Key{}
	if err := key.FromObject(ko); err != nil {
		return err
	}

	hash, err := encoding.ObjectHash(o)
	if err != nil {
		return err
	}

	switch k := key.Materialize().(type) {
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
