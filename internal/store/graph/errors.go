package graph

import (
	"nimona.io/pkg/errors"
)

const (
	ErrMissingID = errors.Error("missing id")
	ErrInternal  = errors.Error("internal error")
	ErrExists    = errors.Error("already exists")
	ErrNotFound  = errors.Error("not found")
)
