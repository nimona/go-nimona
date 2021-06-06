package objectstore

import (
	"time"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/value"
)

const (
	// ErrNotFound is returned when a requested object or cid is not found
	ErrNotFound = errors.Error("not found")
)

//go:generate mockgen -destination=../objectstoremock/objectstoremock_generated.go -package=objectstoremock -source=objectstore.go

type (
	Getter interface {
		Get(cid value.CID) (*object.Object, error)
	}
	Store interface {
		Get(cid value.CID) (*object.Object, error)
		GetByType(string) (object.ReadCloser, error)
		GetByStream(value.CID) (object.ReadCloser, error)
		Put(*object.Object) error
		PutWithTTL(*object.Object, time.Duration) error
		GetStreamLeaves(streamRootCID value.CID) ([]value.CID, error)
		Pin(value.CID) error
		IsPinned(value.CID) (bool, error)
		GetPinned() ([]value.CID, error)
		RemovePin(value.CID) error
	}
)
