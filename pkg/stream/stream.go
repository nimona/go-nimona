package stream

import (
	"fmt"
	"reflect"
	"sync"

	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"

	"nimona.io/pkg/tilde"
)

type (
	NilState struct{}
	Object   object.Object
)

func (e *Object) Apply(s *NilState) error {
	return nil
}

type (
	Applicable[State any] interface {
		Apply(*State) error
	}
	Manager[State any] interface {
		NewStreamController() Controller[State]
		GetStreamController(tilde.Digest) Controller[State]
	}
	Controller[State any] interface {
		ApplyObject(*object.Object) error
		ApplyEvent(Applicable[State]) error
		GetStream() Stream
		GetState() State
	}
	Stream             struct{}
	manager[State any] struct {
		Network          network.Network
		ObjectStore      *sqlobjectstore.Store
		ApplicableEvents map[string]reflect.Type
	}
	controller[State any] struct {
		lock        sync.RWMutex
		LatestState *State
		Metadata    map[tilde.Digest]object.Metadata
	}
)

func NewManager[State any](
	network network.Network,
	objectStore *sqlobjectstore.Store,
	events ...Applicable[State],
) (Manager[State], error) {
	m := &manager[State]{
		Network:          network,
		ObjectStore:      objectStore,
		ApplicableEvents: map[string]reflect.Type{},
	}
	for _, e := range events {
		err := m.registerEvent(e)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *manager[State]) registerEvent(e Applicable[State]) error {
	t := reflect.TypeOf(e)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("expected ptr, got %s", t.Kind())
	}
	t = t.Elem()
	o, err := object.Marshal(e)
	if err != nil {
		return fmt.Errorf("event cannot be marshalled into json, %w", err)
	}
	if o.Type == "" {
		return fmt.Errorf("event does not have a type")
	}
	m.ApplicableEvents[o.Type] = reflect.TypeOf(e)
	return nil
}

func (m *manager[State]) NewStreamController() Controller[State] {
	c := &controller[State]{
		LatestState: new(State),
	}
	return c
}

func (m *manager[State]) GetStreamController(h tilde.Digest) Controller[State] {
	c := &controller[State]{
		LatestState: new(State),
	}
	return c
}

func NewController[State any]() (Controller[State], error) {
	c := &controller[State]{}
	return c, nil
}

func (s *controller[State]) ApplyObject(o *object.Object) error {
	return nil
}

func (s *controller[State]) ApplyEvent(e Applicable[State]) error {
	err := e.Apply(s.LatestState)
	if err != nil {
		return fmt.Errorf("error applying state, %w", err)
	}
	return nil
}

func (s *controller[State]) GetStream() Stream {
	return Stream{}
}

func (s *controller[State]) GetState() State {
	return *s.LatestState
}
