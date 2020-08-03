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

//go:generate $GOBIN/mockgen -destination=../objectstoremock/objectstoremock_generated.go -package=objectstoremock -source=objectstore.go

type (
	Getter interface {
		Get(hash object.Hash) (object.Object, error)
	}
	Store interface {
		Get(hash object.Hash) (object.Object, error)
		GetByType(string) ([]object.Object, error)
		GetByStream(object.Hash) ([]object.Object, error)
		Put(object.Object) error
		PutWithTimeout(object.Object, time.Duration) error
	}
)
