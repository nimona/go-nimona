package object

import (
	"nimona.io/pkg/errors"
)

const (
	// ErrCouldNotVerify is returned when the signature doesn't matches the
	// given key
	ErrCouldNotVerify = errors.Error("could not verify signature")
)

// Verify object
func Verify(o Object) error {
	sig := o.GetSignature()
	if sig.IsEmpty() {
		// TODO return error or nil?
		return nil
	}

	if err := sig.Signer.Verify(
		o.Hash().rawBytes(),
		sig.X,
	); err != nil {
		return err
	}

	return nil
}
