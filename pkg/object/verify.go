package object

import (
	"fmt"

	"nimona.io/pkg/did"
	"nimona.io/pkg/errors"
)

const (
	ErrInvalidSigner    = errors.Error("signer does not match owner")
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
	if sig.IsEmpty() && own == did.Empty {
		return nil
	}

	// if there is an owner, we should have a signature
	if sig.IsEmpty() {
		return ErrMissingSignature
	}

	// get object map
	m, err := o.MarshalMap()
	if err != nil {
		return err
	}

	// get the hash
	h, err := m.Hash().Bytes()
	if err != nil {
		return fmt.Errorf("unable to get bytes from hash, %w", err)
	}

	// verify the signature
	if err := sig.Key.Verify(
		h,
		sig.X,
	); err != nil {
		return err
	}

	// if there is no owner, we're fine
	if own == did.Empty {
		return nil
	}

	// check if the owner matches the signer
	if own == sig.Key.DID() {
		return nil
	}

	// or, error out
	return ErrInvalidSigner
}
