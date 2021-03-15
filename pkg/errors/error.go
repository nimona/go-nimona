package errors

import (
	"errors"
	"fmt"
)

type (
	Error string
	// wError is a wrapper to allow nesting errors
	// Holds two errors, a latest and a cause
	wError struct {
		outer error
		inner error
	}
)

// Error prints the error's message
func (e Error) Error() string {
	return string(e)
}

// Error prints the message error, appended with the underlying cause
func (err wError) Error() string {
	if err.inner != nil {
		return fmt.Sprintf("%s: %v", err.outer, err.inner)
	}
	return err.outer.Error()
}

func (err wError) Unwrap() error {
	return err.inner
}

func (err wError) Is(target error) bool {
	if errors.Is(err.inner, target) {
		return true
	}
	return errors.Is(err.outer, target)
}

// Merge two errors
func Merge(outer, inner error) error {
	return wError{
		outer: outer,
		inner: inner,
	}
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
