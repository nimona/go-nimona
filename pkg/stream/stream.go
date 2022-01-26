package stream

import (
	"fmt"
	"reflect"

	"nimona.io/pkg/object"
)

type (
	// ApplyFn[State any]    func(State) error
	Applicable[State any] interface {
		// GetType() string
		Apply(*State) error
	}
	Controller[State any] interface {
		// RegisterHandler(string, ApplyFn[State]) error
		// RegisterEvent(Applicable[State]) error
		NewStream() Stream[State]
	}
	Stream[State any] interface {
		ApplyObject(*object.Object) error
		ApplyEvent(Applicable[State]) error
		GetState() State
	}
	controller[State any] struct {
		ApplicableEvents map[string]reflect.Type
	}
	stream[State any] struct {
		// lock sync.RWMutex // TODO add mutex
		Controller  Controller[State]
		LatestState *State
	}
)

func NewController[State any](
	events ...Applicable[State],
) (Controller[State], error) {
	c := &controller[State]{
		ApplicableEvents: map[string]reflect.Type{},
	}
	for _, e := range events {
		err := c.registerEvent(e)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *controller[State]) registerEvent(e Applicable[State]) error {
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
	c.ApplicableEvents[o.Type] = reflect.TypeOf(e)
	return nil
}

func (c *controller[State]) NewStream() Stream[State] {
	s := &stream[State]{
		LatestState: new(State),
		Controller:  c,
	}
	return s
}

func (s *stream[State]) ApplyObject(o *object.Object) error {
	return nil
}

func (s *stream[State]) ApplyEvent(e Applicable[State]) error {
	err := e.Apply(s.LatestState)
	if err != nil {
		return fmt.Errorf("error applying state, %w", err)
	}
	return nil
}

func (s *stream[State]) GetState() State {
	return *s.LatestState
}
