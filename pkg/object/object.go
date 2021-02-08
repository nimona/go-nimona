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

func (o Object) ToMap() Map {
	return o.Map()
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

func (o *Object) Hash() Hash {
	h, err := NewHash(o)
	if err != nil {
		panic(err)
	}
	return h
}

func (h Hash) String() string {
	return string(h)
}
