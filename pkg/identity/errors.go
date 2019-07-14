package identity

import "errors"

var (
	// ErrMissingKey when a key is not passed
	ErrMissingKey = errors.New("missing key")
	// ErrECDSAPrivateKeyRequired when a key is not an ECDSA key
	ErrECDSAPrivateKeyRequired = errors.New(
		"network currently requires an ecdsa private key")
)
