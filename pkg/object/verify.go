package object

import (
	"nimona.io/pkg/errors"
)

const (
	ErrOwnerDoesNotMatchSigner = errors.Error("owner does not match signer")
	// ErrCouldNotVerify is returned when the signature doesn't matches the
	// given key
	ErrCouldNotVerify = errors.Error("could not verify signature")
)

// Verify object
// TODO should this verify nested objects as well?
func Verify(o Object) error {
	sig := o.GetSignature()
	if sig.IsEmpty() {
		return nil
	}

	if !o.GetOwner().IsEmpty() && o.GetOwner() != sig.Signer {
		return ErrOwnerDoesNotMatchSigner
	}

	if err := sig.Signer.Verify(
		o.Hash().rawBytes(),
		sig.X,
	); err != nil {
		return err
	}

	return nil
}
