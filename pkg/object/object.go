package object

import (
	"encoding/json"

	"nimona.io/pkg/errors"
)

const (
	ErrSourceNotSupported = errors.Error("encoding source not supported")
)

type (
	Typed interface {
		Type() string
	}
	// Object
	Object struct {
		Type     string
		Metadata Metadata
		Data     Map
	}
)

// TODO: Deprecate
func (o Object) Map() Map {
	r := Map{}
	if o.Type != "" {
		r["type"] = String(o.Type)
	}
	mm := o.Metadata.Map()
	if len(mm) > 0 {
		r["metadata"] = o.Metadata.Map()
	}
	if len(o.Data) > 0 {
		r["data"] = o.Data
	}
	return r
}

// TODO: Deprecate
func (o Object) ToMap() Map {
	return o.Map()
}

func (o Object) MarshalObject() (*Object, error) {
	return &o, nil
}

func (o Object) MarshalMap() (Map, error) {
	return o.Map(), nil
}

func (o *Object) UnmarshalJSON(b []byte) error {
	m := Map{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if t, ok := m["type"]; ok {
		if s, ok := t.(String); ok {
			o.Type = string(s)
		}
	}
	if t, ok := m["metadata"]; ok {
		if s, ok := t.(Map); ok {
			o.Metadata = MetadataFromMap(s)
		}
	}
	if t, ok := m["data"]; ok {
		if s, ok := t.(Map); ok {
			o.Data = s
		}
	}
	return nil
}

func (o Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Map())
}

func (o *Object) UnmarshalObject(n *Object) error {
	o.Type = n.Type
	o.Data = n.Data
	o.Metadata = n.Metadata
	return nil
}

func (o *Object) UnmarshalMap(m Map) error {
	if t, ok := m["type"]; ok {
		if s, ok := t.(String); ok {
			o.Type = string(s)
		}
	}
	if t, ok := m["metadata"]; ok {
		if s, ok := t.(Map); ok {
			o.Metadata = MetadataFromMap(s)
		}
	}
	if t, ok := m["data"]; ok {
		if s, ok := t.(Map); ok {
			o.Data = s
		}
	}
	return nil
}

// TODO: Deprecate
func FromMap(m Map) *Object {
	o := &Object{}
	if t, ok := m["type"]; ok {
		if s, ok := t.(String); ok {
			o.Type = string(s)
		}
	}
	if t, ok := m["metadata"]; ok {
		if s, ok := t.(Map); ok {
			o.Metadata = MetadataFromMap(s)
		}
	}
	if t, ok := m["data"]; ok {
		if s, ok := t.(Map); ok {
			o.Data = s
		}
	}
	return o
}

func (o *Object) CID() CID {
	if o == nil {
		return EmptyCID
	}
	h, err := NewCID(o)
	if err != nil {
		panic(err)
	}
	return h
}

func (h CID) String() string {
	return string(h)
}
