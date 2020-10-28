package objectstore

import (
	"time"

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
		Get(hash object.Hash) (*object.Object, error)
	}
	Store interface {
		Get(hash object.Hash) (*object.Object, error)
		GetByType(string) (object.ReadCloser, error)
		GetByStream(object.Hash) (object.ReadCloser, error)
		Put(*object.Object) error
		// TODO rename to PutWithTTL
		PutWithTimeout(*object.Object, time.Duration) error
		// TODO GetPinned should be replaced with something "better"
		GetPinned() ([]object.Hash, error)
	}
)
