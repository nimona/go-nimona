package stream

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

const (
	ErrNotFound = errors.Error("not found")
)

type (
	Manager interface {
		GetOrCreateController(tilde.Digest) (Controller, error)
		GetController(tilde.Digest) (Controller, error)
		Fetch(context.Context, Controller, tilde.Digest) (int, error)
		// Sync(context.Context, tilde.Digest) error
	}
	SyncStrategy interface {
		Fetch(context.Context, Controller, tilde.Digest) (int, error)
		Serve(context.Context, Manager)
	}
	Controller interface {
		Apply(interface{}) error
		Insert(interface{}) (tilde.Digest, error)
		GetStreamInfo() Info
		GetStreamRoot() tilde.Digest
		GetDigests() ([]tilde.Digest, error)
		GetSubscribers() ([]peer.ID, error)
		ContainsDigest(cid tilde.Digest) bool
		GetReader(context.Context) (object.ReadCloser, error)
		// Sync(context.Context) error
		// Subscribe(context.Context) (object.ReadCloser, error)
	}
)

type (
	Applicable[State any] interface {
		Apply(*State) error
	}
	StatefulManager[State any] interface {
		NewController() StatefulController[State]
		GetController(tilde.Digest) StatefulController[State]
	}
	StatefulController[State any] interface {
		Apply(Applicable[State]) error
		GetStreamInfo() Info
		GetStreamRoot() tilde.Digest
		GetStreamState() State
	}
)
