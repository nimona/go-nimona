package objectstore

import (
	"time"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
)

const (
	// ErrNotFound is returned when a requested object or cid is not found
	ErrNotFound = errors.Error("not found")
)

//go:generate mockgen -destination=../objectstoremock/objectstoremock_generated.go -package=objectstoremock -source=objectstore.go

type (
	Getter interface {
		Get(cid object.CID) (*object.Object, error)
	}
	Store interface {
		Get(cid object.CID) (*object.Object, error)
		GetByType(string) (object.ReadCloser, error)
		GetByStream(object.CID) (object.ReadCloser, error)
		Put(*object.Object) error
		PutWithTTL(*object.Object, time.Duration) error
		GetStreamLeaves(streamRootCID object.CID) ([]object.CID, error)
		Pin(object.CID) error
		IsPinned(object.CID) (bool, error)
		GetPinned() ([]object.CID, error)
		RemovePin(object.CID) error
	}
)
