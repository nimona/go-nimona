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
	sigs := o.GetSignatures()
	if len(sigs) == 0 {
		return errors.New("missing signature")
	}

	h := NewHash(o)
	for _, s := range sigs {
		if err := s.Signer.Verify(h.Bytes(), s.X); err != nil {
			return err
		}
	}

	return nil
}
