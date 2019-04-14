// Code generated by nimona.io/tools/objectify. DO NOT EDIT.

// +build !generate

package telemetry

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/pkg/object"
)

const (
	ConnectionEventType = "nimona.io/telemetry/connection"
)

// ToObject returns a f12n object
func (s ConnectionEvent) ToObject() *object.Object {
	o := object.New()
	o.SetType(ConnectionEventType)
	if s.Direction != "" {
		o.SetRaw("direction", s.Direction)
	}
	return o
}

func anythingToAnythingForConnectionEvent(
	from interface{},
	to interface{},
) error {
	config := &mapstructure.DecoderConfig{
		Result:  to,
		TagName: "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(from); err != nil {
		return err
	}

	return nil
}

// FromObject populates the struct from a f12n object
func (s *ConnectionEvent) FromObject(o *object.Object) error {
	atoa := anythingToAnythingForConnectionEvent
	if err := atoa(o.GetRaw("direction"), &s.Direction); err != nil {
		return err
	}

	return nil
}

// GetType returns the object's type
func (s ConnectionEvent) GetType() string {
	return ConnectionEventType
}
