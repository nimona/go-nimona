package object

import (
	"fmt"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/object/cid"
)

const (
	ErrInvalidSigner    = errors.Error("signer is does not match owner")
	ErrMissingSignature = errors.Error("missing signature")
	ErrCouldNotVerify   = errors.Error("could not verify signature")
)

// Verify object
// TODO should this verify nested objects as well?
func Verify(o *Object) error {
	if o == nil {
		return errors.Error("no object")
	}
	sig := o.Metadata.Signature
	own := o.Metadata.Owner

	// if there is no owner and no signature, we're fine
	if sig.IsEmpty() && own.IsEmpty() {
		return nil
	}

	// if there is an owner, we should have a signature
	if sig.IsEmpty() {
		return ErrMissingSignature
	}

	// get object cid
	m, err := o.MarshalMap()
	if err != nil {
		return err
	}
	h, err := cid.New(m)
	if err != nil {
		return err
	}

	// verify the signature
	if err := sig.Signer.Verify(
		[]byte(h),
		sig.X,
	); err != nil {
		return err
	}

	// if there is no owner, we're fine
	if own.IsEmpty() {
		return nil
	}

	// check if the owner matches the signer
	if own.Equals(sig.Signer) {
		return nil
	}

	// or that the signature contains a valid certificate signed by the owner
	// check if there is a certifiate
	if sig.Certificate == nil {
		return ErrInvalidSigner
	}

	// then let's make sure that the certificate is properly signed
	co, err := sig.Certificate.MarshalObject()
	if err != nil {
		return err
	}
	if err := Verify(co); err != nil {
		return fmt.Errorf("error verifying certificate, %w", err)
	}

	// finally check that the certificate signer matches the object owner
	if !sig.Certificate.Metadata.Signature.Signer.Equals(own) {
		return ErrInvalidSigner
	}

	// and the certificate subject matches the object signer
	if sig.Certificate.Subject.Equals(sig.Signer) {
		return nil
	}

	// or, error out
	return ErrInvalidSigner
}
