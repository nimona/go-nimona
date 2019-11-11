package crypto

import (
	"nimona.io/pkg/errors"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
)

const (
	// ErrCouldNotVerify is returned when the signature doesn't matches the
	// given key
	ErrCouldNotVerify = errors.Error("could not verify signature")
)

// Verify object
func Verify(o object.Object) error {
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

	h := hash.New(o)
	return sig.Signer.Subject.Verify(h.D, sig.X)
}
