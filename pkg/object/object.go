package object

import (
	"encoding/json"
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

// TODO: Deprecate
func (o Object) Map() Map {
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
	mm := o.Metadata.Map()
	if len(mm) > 0 {
		m["@metadata"] = o.Metadata.Map()
	}
	return m
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
	return o.UnmarshalMap(m)
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
			o.Metadata = MetadataFromMap(s)
			delete(m, "@metadata")
		}
	}
	o.Data = m
	return nil
}

// TODO: Deprecate
func FromMap(m Map) *Object {
	o := &Object{}
	o.UnmarshalMap(m) // nolint: errcheck
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
