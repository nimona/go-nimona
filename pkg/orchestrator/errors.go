package orchestrator

import (
	"nimona.io/pkg/errors"
)

const (
	// ErrIncompleteGraph is returned when a requested graph is incomplete
	ErrIncompleteGraph = errors.Error("incomplete graph")
)
