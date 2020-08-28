package network

import (
	"nimona.io/pkg/errors"
)

var (
	// ErrNotFound is returned by Get() when the object was not found
	ErrNotFound = errors.New("not found")
	// ErrCannotSendToSelf is returned when trying to Send() to our own peer
	ErrCannotSendToSelf = errors.New("cannot send objects to ourself")
)
