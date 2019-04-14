package errors

type (
	Error string
	// iError augments the error interface with helpers for our wrapped errors
	iError interface {
		error
		Cause() error
		Latest() error
	}
	// fError is a fundamental error with just a message
	// TODO(geoah): consider adding a stack as pkg/errors does
	fError struct {
		message string
	}
	// wError is a wrapper to allow nesting errors
	// Holds two errors, a latest and a cause
	wError struct {
		error error
		cause error
	}
)

// Error prints the error's message
func (e Error) Error() string {
	return string(e)
}

// iError prints the message error, appended with the underlying cause
func (e fError) Error() string {
	return e.message
}

// iError prints the message error, appended with the underlying cause
func (e wError) Error() string {
	if e.cause == nil {
		return e.error.Error()
	}
	return e.error.Error() + ": " + e.cause.Error()
}

// Cause returns the next error in the error chain.
// If there is no next error, Cause returns nil.
func (e wError) Cause() error {
	return e.cause
}

// Latest returns the last error in its original form
func (e wError) Latest() error {
	return e.error
}

// New returns an error with the supplied message.
func New(message string) error {
	return fError{
		message: message,
	}
}

// Wrap an error with a cause
func Wrap(err, cause error) error {
	return wError{
		cause: cause,
		error: err,
	}
}

// Unwrap returns the next error in the error chain.
// If there is no next error, Unwrap returns nil.
func Unwrap(err error) error {
	if e, ok := err.(interface{ Cause() error }); ok {
		return e.Cause()
	}
	return nil
}

// CausedBy checks if an error was caused by another
func CausedBy(err, cause error) bool {
	if err == nil {
		return false
	}
	if err == cause {
		return true
	}
	if e, ok := err.(wError); ok {
		return CausedBy(cause, e.error) || CausedBy(cause, e.cause)
	}
	return false
}
