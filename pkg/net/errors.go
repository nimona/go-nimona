package net

import "errors"

var (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
	// ErrNoAddresses for when a peer has no addresses
	ErrNoAddresses = errors.New("no addresses")
	// ErrNotForUs object is not meant for us
	ErrNotForUs = errors.New("object not for us")
	// ErrMissingKey when a key is not passed
	ErrMissingKey = errors.New("missing key")
	// ErrPrivateKeyRequired when a key is not an ECDSA key
	ErrECDSAPrivateKeyRequired = errors.New(
		"network currently requires an ecdsa private key")
)
