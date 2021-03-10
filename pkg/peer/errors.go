package peer

import "nimona.io/pkg/errors"

const (
	// ErrMissingKey when a key is not passed
	ErrMissingKey = errors.Error("missing key")
	// ErrECDSAPrivateKeyRequired when a key is not an ECDSA key
	ErrECDSAPrivateKeyRequired = errors.Error(
		"network currently requires an ecdsa private key")
)
