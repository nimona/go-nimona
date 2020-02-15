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
	sig := o.Header.Signature
	if sig.IsEmpty() {
		return errors.New("missing signature")
	}

	h := NewHash(o)
	return sig.Signer.Verify(h.Bytes(), sig.X)
}
