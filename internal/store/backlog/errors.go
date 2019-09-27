package backlog

import (
	"nimona.io/pkg/errors"
)

var (
	// ErrNoMoreObjects is returned on Pop() when there are no more objects
	// to return
	ErrNoMoreObjects = errors.New("no more objects")
	// ErrAlreadyExists is returned on Push() when combination of object/key
	// already exists
	ErrAlreadyExists = errors.New("already exists")
)
