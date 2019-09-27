package kv

import (
	"nimona.io/pkg/errors"
)

var (
	// ErrNotFound ...
	ErrNotFound = errors.New("not found")
	ErrEmpty    = errors.New("empty")
	ErrExists   = errors.New("already exists")
)
