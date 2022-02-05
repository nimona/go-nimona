package stream

import (
	"reflect"
	"sync"

	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

type (
	Manager interface {
		NewStreamController() Controller
		GetStreamController(tilde.Digest) Controller
	}
	Controller interface {
		Apply(interface{}) (tilde.Digest, error)
		GetStreamInfo() StreamInfo
		GetStreamRoot() tilde.Digest
	}
)

type (
	Applicable[State any] interface {
		Apply(*State) error
	}
	StatefulManager[State any] interface {
		NewStreamController() StatefulController[State]
		GetStreamController(tilde.Digest) StatefulController[State]
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
