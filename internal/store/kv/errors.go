package kv

import (
	"nimona.io/internal/errors"
)

var (
	// ErrNotFound ...
	ErrNotFound = errors.New("not found")
	ErrEmpty    = errors.New("empty")
	ErrExists   = errors.New("already exists")
)
