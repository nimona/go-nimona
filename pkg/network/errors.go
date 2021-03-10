package network

import (
	"nimona.io/pkg/errors"
)

const (
	// ErrNotFound is returned by Get() when the object was not found
	ErrNotFound = errors.Error("not found")
	// ErrCannotSendToSelf is returned when trying to Send() to our own peer
	ErrCannotSendToSelf = errors.Error("cannot send objects to ourself")
	// ErrInvalidRequest when received an invalid request object
	ErrInvalidRequest = errors.Error("invalid request")
	// ErrSendingTimedOut when sending times out
	ErrSendingTimedOut = errors.Error("sending timed out")
	// ErrAlreadySentDuringContext when trying to send to the same peer during
	// this context
	ErrAlreadySentDuringContext = errors.Error("already sent to peer")
	// ErrWaitingForResponseTimedOut is returned when Send is waiting for a
	// response given a response id
	ErrWaitingForResponseTimedOut = errors.Error("time out waiting for response")
	// ErrUnableToUnmarshalIntoResponse is returned when the returned object
	// cannot be unmarshalled into given struct
	ErrUnableToUnmarshalIntoResponse = errors.Error("unable to unmarshal into" +
		" given response")
)
