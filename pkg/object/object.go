package object

import (
	"encoding/json"
	"reflect"

	"github.com/mitchellh/copystructure"

	"nimona.io/pkg/tilde"
)

type Object struct {
	Context  tilde.Digest
	Type     string
	Metadata Metadata
	Data     tilde.Map
}

type (
	Typer interface {
		Type() string
	}
	MapMashaller interface {
		MarshalMap() (tilde.Map, error)
	}
	MapUnmashaller interface {
		UnmarshalMap(tilde.Map) error
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
	m := tilde.Map{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	return o.UnmarshalMap(m)
}

func (o *Object) MarshalMap() (tilde.Map, error) {
	m := tilde.Map{}
	for k, v := range o.Data {
		m[k] = v
	}
	if o.Context != "" {
		m["@context"] = tilde.String(o.Context)
	}
	if o.Type != "" {
		m["@type"] = tilde.String(o.Type)
	}
	mm, err := marshalStruct(tilde.MapHint, reflect.ValueOf(o.Metadata))
	if err != nil {
		return nil, err
	}
	if len(mm) > 0 {
		m["@metadata"] = mm
	}
	return m, nil
}

func (o *Object) UnmarshalMap(v tilde.Map) error {
	mm, err := copystructure.Copy(v)
	if err != nil {
		return err
	}
	m := mm.(tilde.Map)
	if t, ok := m["@context"]; ok {
		if s, ok := t.(tilde.String); ok {
			o.Context = tilde.Digest(s)
			delete(m, "@context")
		}
	}
	if t, ok := m["@type"]; ok {
		if s, ok := t.(tilde.String); ok {
			o.Type = string(s)
			delete(m, "@type")
		}
	}
	if t, ok := m["@metadata"]; ok {
		if s, ok := t.(tilde.Map); ok {
			err := unmarshalMapToStruct(
				tilde.MapHint,
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

func (o *Object) Hash() tilde.Digest {
	if o == nil {
		return tilde.EmptyDigest
	}
	m, err := o.MarshalMap()
	if err != nil {
		panic("object.Hash(), MarshalMap should not error")
	}
	return m.Hash()
}
