package object

import (
	"encoding/json"
	"reflect"

	"github.com/mitchellh/copystructure"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/object/cid"
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
		Data     chore.Map
	}
)

// nolint: golint
type (
	ObjectMashaller interface {
		MarshalObject() (*Object, error)
	}
	ObjectUnmashaller interface {
		UnmarshalObject(*Object) error
	}
	MapMashaller interface {
		MarshalMap() (chore.Map, error)
	}
	MapUnmashaller interface {
		UnmarshalMap(chore.Map) error
	}
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

func (o *Object) MarshalJSON() ([]byte, error) {
	m, err := o.MarshalMap()
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func (o *Object) UnmarshalJSON(b []byte) error {
	m := chore.Map{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	return o.UnmarshalMap(m)
}

func (o *Object) MarshalMap() (chore.Map, error) {
	m := chore.Map{}
	for k, v := range o.Data {
		m[k] = v
	}
	if o.Context != "" {
		m["@context"] = chore.String(o.Context)
	}
	if o.Type != "" {
		m["@type"] = chore.String(o.Type)
	}
	mm, err := marshalStruct(chore.MapHint, reflect.ValueOf(o.Metadata))
	if err != nil {
		return nil, err
	}
	if len(mm) > 0 {
		m["@metadata"] = mm
	}
	return m, nil
}

func (o *Object) UnmarshalMap(v chore.Map) error {
	mm, err := copystructure.Copy(v)
	if err != nil {
		return err
	}
	m := mm.(chore.Map)
	if t, ok := m["@context"]; ok {
		if s, ok := t.(chore.String); ok {
			o.Context = string(s)
			delete(m, "@context")
		}
	}
	if t, ok := m["@type"]; ok {
		if s, ok := t.(chore.String); ok {
			o.Type = string(s)
			delete(m, "@type")
		}
	}
	if t, ok := m["@metadata"]; ok {
		if s, ok := t.(chore.Map); ok {
			err := unmarshalMapToStruct(
				chore.MapHint,
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

// TODO also return error
func (o *Object) CID() chore.CID {
	if o == nil {
		return cid.Empty
	}
	m, err := o.MarshalMap()
	if err != nil {
		return cid.Invalid
	}
	h, err := cid.New(m)
	if err != nil {
		return cid.Invalid
	}
	return h
}
