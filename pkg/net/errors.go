package net

import "nimona.io/pkg/errors"

const (
	// ErrAllAddressesFailed for when a peer cannot be dialed
	ErrAllAddressesFailed = errors.Error("all addresses failed to dial")
	// ErrNoAddresses for when a peer has no addresses
	ErrNoAddresses = errors.Error("no addresses")
	// ErrMissingSignature is when the signature is missing
	ErrMissingSignature = errors.Error("signature missing")
	// ErrAllAddressesBlocked all peer's addresses are currently blocked
	ErrAllAddressesBlocked = errors.Error("all addresses blocked")
	// ErrInvalidSignature signature is invalid
	ErrInvalidSignature = errors.Error("invalid signature")
	// ErrConnectionClosed connection is closed, will usually be merged with
	// an underlying error
	ErrConnectionClosed = errors.Error("connection closed")
)
