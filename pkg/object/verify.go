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
	if o == nil {
		return errors.New("missing object")
	}

	sig, err := GetObjectSignature(o)
	if err != nil {
		return errors.Wrap(
			errors.New("could not get signature"),
			err,
		)
	}

	h := NewHash(o)
	return sig.Signer.Verify(h.Bytes(), sig.X)
}
