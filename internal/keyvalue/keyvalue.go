// Package keyvalue implements an in-memory key-value store.

package keyvalue

import (
	"nimona.io/pkg/errors"
)

const (
	// ErrNotFound is returned when the key is not found.
	ErrNotFound = errors.Error("key not found")
)

type Store interface {
	// Get gets the value for the given key.
	// If the key does not exist, then Get returns ErrNotFound.
	Get(key string) (value []byte, err error)
	// Set sets the value for the given key.
	Set(key string, value []byte) error
	// Delete deletes the value for the given key.
	Delete(key string) error
	// Iter calls f for each key-value pair in the store.
	Iter(f func(key string, value []byte))
	// Close closes the store.
	Close() error
}
