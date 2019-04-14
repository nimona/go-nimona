package dag

import (
	"nimona.io/internal/errors"
)

const (
	// ErrIncompleteGraph is returned when a requested graph is incomplete
	ErrIncompleteGraph = errors.Error("incomplete graph")
)
