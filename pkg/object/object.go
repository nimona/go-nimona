package object

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type (
	Typed interface {
		Type() string
	}
	// Object
	Object struct {
		Context  string
		Type     string
		Metadata Metadata
		Data     Map
	}
)

type (
	StringMashaller interface {
		MarshalString() (string, error)
	}
	StringUnmashaller interface {
		UnmarshalString(string) error
	}
	ByteUnmashaller interface {
		UnmarshalBytes([]byte) error
	}
	ByteMashaller interface {
		MarshalBytes() ([]byte, error)
	}
)

func (o *Object) MarshalMap() (Map, error) {
	m := Map{}
	for k, v := range o.Data {
		m[k] = v
	}
	if o.Context != "" {
		m["@context"] = String(o.Context)
	}
	if o.Type != "" {
		m["@type"] = String(o.Type)
	}
	mm, err := marshalStruct(MapHint, reflect.ValueOf(o.Metadata))
	if err != nil {
		return nil, err
	}
	if len(mm) > 0 {
		m["@metadata"] = mm
	}
	return m, nil
}

func (o *Object) UnmarshalJSON(b []byte) error {
	m := Map{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	return o.UnmarshalMap(m)
}

func (o *Object) MarshalJSON() ([]byte, error) {
	m, err := o.MarshalMap()
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func (o *Object) UnmarshalMap(m Map) error {
	if t, ok := m["@context"]; ok {
		if s, ok := t.(String); ok {
			o.Context = string(s)
			delete(m, "@context")
		}
	}
	if t, ok := m["@type"]; ok {
		if s, ok := t.(String); ok {
			o.Type = string(s)
			delete(m, "@type")
		}
	}
	if t, ok := m["@metadata"]; ok {
		if s, ok := t.(Map); ok {
			err := unmarshalMapToStruct(
				MapHint,
				s,
				reflect.ValueOf(&o.Metadata),
			)
			if err != nil {
				return err
			}
			delete(m, "@metadata")
		}
	}
	o.Data = m
	return nil
}

func (o *Object) CID() CID {
	if o == nil {
		return EmptyCID
	}
	h, err := NewCID(o)
	if err != nil {
		panic(fmt.Errorf("object.CID() panicked: %w", err))
	}
	return h
}

func (h CID) String() string {
	return string(h)
}
