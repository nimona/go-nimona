package object

import (
	"nimona.io/pkg/errors"
)

const (
	ErrInvalidSigner    = errors.Error("signer is does not match owner")
	ErrMissingSignature = errors.Error("missing signature")
	ErrCouldNotVerify   = errors.Error("could not verify signature")
)

// Verify object
// TODO should this verify nested objects as well?
func Verify(o Object) error {
	sig := o.GetSignature()
	own := o.GetOwner()

	// if there is no owner and no signature, we're fine
	if sig.IsEmpty() && own.IsEmpty() {
		return nil
	}

	// if there is an owner, we should have a signature
	if sig.IsEmpty() {
		return ErrMissingSignature
	}

	// verify the signature
	if err := sig.Signer.Verify(
		o.Hash().rawBytes(),
		sig.X,
	); err != nil {
		return err
	}

	// if there is no owner, we're fine
	if own.IsEmpty() {
		return nil
	}

	// check if the owner matches the signer
	if own == sig.Signer {
		return nil
	}

	// or that the signature contains a valid certificate signed by the owner
	// check if there is a certifiate
	if sig.Certificate == nil {
		return ErrInvalidSigner
	}

	// then let's make sure that the certificate is properly signed
	if err := Verify(sig.Certificate.ToObject()); err != nil {
		return errors.Wrap(
			errors.New("error verifying certificate"),
			err,
		)
	}

	// finally check that the certificate signer matches the object owner
	if sig.Certificate.Metadata.Signature.Signer != own {
		return ErrInvalidSigner
	}

	// and the certificate subject matches the object signer
	for _, sub := range sig.Certificate.Metadata.Policy.Subjects {
		if sub == sig.Signer.String() {
			return nil
		}
	}

	// or, error out
	return ErrInvalidSigner
}
