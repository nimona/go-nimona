package exchange

import (
	"nimona.io/internal/errors"
)

var (
	// ErrNotFound is returned by Get() when the object was not found
	ErrNotFound = errors.New("not found")
)
