package stream

import (
	"reflect"
	"sync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

type (
	Manager interface {
		NewController() Controller
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
		GetStreamInfo() StreamInfo
		GetStreamRoot() tilde.Digest
		GetDigests() ([]tilde.Digest, error)
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
		GetStreamInfo() StreamInfo
		GetStreamRoot() tilde.Digest
		GetStreamState() State
	}
	statefulManager[State any] struct {
		Network          network.Network
		ObjectStore      *sqlobjectstore.Store
		ApplicableEvents map[string]reflect.Type
	}
	statefulController[State any] struct {
		lock        sync.RWMutex
		LatestState *State
		Metadata    map[tilde.Digest]object.Metadata
	}
)
