package crypto

import "nimona.io/pkg/errors"

const (
	ErrUnsupportedKeyAlgorithm = errors.Error("key algorithm not supported")
	ErrInvalidSignature        = errors.Error("invalid signature")
)
