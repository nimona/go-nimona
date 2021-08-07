package objectstore

import (
	"time"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
	"nimona.io/pkg/tilde"
)

const (
	// ErrNotFound is returned when a requested object or hash is not found
	ErrNotFound = errors.Error("not found")
)

//go:generate mockgen -destination=../objectstoremock/objectstoremock_generated.go -package=objectstoremock -source=objectstore.go

type (
	Getter interface {
		Get(hash tilde.Hash) (*object.Object, error)
	}
	Store interface {
		Get(hash tilde.Hash) (*object.Object, error)
		GetByType(string) (object.ReadCloser, error)
		GetByStream(tilde.Hash) (object.ReadCloser, error)
		Put(*object.Object) error
		PutWithTTL(*object.Object, time.Duration) error
		GetStreamLeaves(streamRootHash tilde.Hash) ([]tilde.Hash, error)
		Pin(tilde.Hash) error
		IsPinned(tilde.Hash) (bool, error)
		GetPinned() ([]tilde.Hash, error)
		RemovePin(tilde.Hash) error
	}
)
