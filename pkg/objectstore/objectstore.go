package objectstore

import (
	"time"

	"nimona.io/pkg/chore"
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
		Get(cid chore.CID) (*object.Object, error)
	}
	Store interface {
		Get(cid chore.CID) (*object.Object, error)
		GetByType(string) (object.ReadCloser, error)
		GetByStream(chore.CID) (object.ReadCloser, error)
		Put(*object.Object) error
		PutWithTTL(*object.Object, time.Duration) error
		GetStreamLeaves(streamRootCID chore.CID) ([]chore.CID, error)
		Pin(chore.CID) error
		IsPinned(chore.CID) (bool, error)
		GetPinned() ([]chore.CID, error)
		RemovePin(chore.CID) error
	}
)
