package objectstore

import (
	"time"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
)

const (
	// ErrNotFound is returned when a requested object or hash is not found
	ErrNotFound = errors.Error("not found")
)

//go:generate mockgen -destination=../objectstoremock/objectstoremock_generated.go -package=objectstoremock -source=objectstore.go

type (
	Getter interface {
		Get(hash chore.Hash) (*object.Object, error)
	}
	Store interface {
		Get(hash chore.Hash) (*object.Object, error)
		GetByType(string) (object.ReadCloser, error)
		GetByStream(chore.Hash) (object.ReadCloser, error)
		Put(*object.Object) error
		PutWithTTL(*object.Object, time.Duration) error
		GetStreamLeaves(streamRootHash chore.Hash) ([]chore.Hash, error)
		Pin(chore.Hash) error
		IsPinned(chore.Hash) (bool, error)
		GetPinned() ([]chore.Hash, error)
		RemovePin(chore.Hash) error
	}
)
