package graph

import (
	"nimona.io/internal/errors"
)

const (
	ErrMissingID = errors.Error("missing id")
	ErrInternal  = errors.Error("internal error")
	ErrExists    = errors.Error("already exists")
	ErrNotFound  = errors.Error("not found")
)
