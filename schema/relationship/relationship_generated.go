// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package relationship

import (
	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

type (
	RelationshipStreamRoot struct {
		Metadata object.Metadata
	}
	Added struct {
		Metadata    object.Metadata
		Alias       string
		RemoteParty crypto.PublicKey
		Datetime    string
	}
	Removed struct {
		Metadata    object.Metadata
		RemoteParty crypto.PublicKey
		Datetime    string
	}
)

func (e *RelationshipStreamRoot) Type() string {
	return "stream:nimona.io/schema/relationship"
}

func (e *RelationshipStreamRoot) MarshalMap() (object.Map, error) {
	return e.ToObject().Map(), nil
}

func (e *RelationshipStreamRoot) MarshalObject() (*object.Object, error) {
	return e.ToObject(), nil
}

func (e RelationshipStreamRoot) ToObject() *object.Object {
	r := &object.Object{
		Type:     "stream:nimona.io/schema/relationship",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	return r
}

func (e *RelationshipStreamRoot) UnmarshalMap(m object.Map) error {
	return e.FromObject(object.FromMap(m))
}

func (e *RelationshipStreamRoot) UnmarshalObject(o *object.Object) error {
	return e.FromObject(o)
}

func (e *RelationshipStreamRoot) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	return nil
}

func (e *Added) Type() string {
	return "event:nimona.io/schema/relationship.Added"
}

func (e *Added) MarshalMap() (object.Map, error) {
	return e.ToObject().Map(), nil
}

func (e *Added) MarshalObject() (*object.Object, error) {
	return e.ToObject(), nil
}

func (e Added) ToObject() *object.Object {
	r := &object.Object{
		Type:     "event:nimona.io/schema/relationship.Added",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	r.Data["alias"] = object.String(e.Alias)
	if v, err := e.RemoteParty.MarshalString(); err == nil {
		r.Data["remoteParty"] = object.String(v)
	}
	r.Data["datetime"] = object.String(e.Datetime)
	return r
}

func (e *Added) UnmarshalMap(m object.Map) error {
	return e.FromObject(object.FromMap(m))
}

func (e *Added) UnmarshalObject(o *object.Object) error {
	return e.FromObject(o)
}

func (e *Added) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["alias"]; ok {
		if t, ok := v.(object.String); ok {
			e.Alias = string(t)
		}
	}
	if v, ok := o.Data["remoteParty"]; ok {
		if ev, ok := v.(object.String); ok {
			es := crypto.PublicKey{}
			if err := es.UnmarshalString(string(ev)); err == nil {
				e.RemoteParty = es
			}
		}
	}
	if v, ok := o.Data["datetime"]; ok {
		if t, ok := v.(object.String); ok {
			e.Datetime = string(t)
		}
	}
	return nil
}

func (e *Removed) Type() string {
	return "event:nimona.io/schema/relationship.Removed"
}

func (e *Removed) MarshalMap() (object.Map, error) {
	return e.ToObject().Map(), nil
}

func (e *Removed) MarshalObject() (*object.Object, error) {
	return e.ToObject(), nil
}

func (e Removed) ToObject() *object.Object {
	r := &object.Object{
		Type:     "event:nimona.io/schema/relationship.Removed",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	if v, err := e.RemoteParty.MarshalString(); err == nil {
		r.Data["remoteParty"] = object.String(v)
	}
	r.Data["datetime"] = object.String(e.Datetime)
	return r
}

func (e *Removed) UnmarshalMap(m object.Map) error {
	return e.FromObject(object.FromMap(m))
}

func (e *Removed) UnmarshalObject(o *object.Object) error {
	return e.FromObject(o)
}

func (e *Removed) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["remoteParty"]; ok {
		if ev, ok := v.(object.String); ok {
			es := crypto.PublicKey{}
			if err := es.UnmarshalString(string(ev)); err == nil {
				e.RemoteParty = es
			}
		}
	}
	if v, ok := o.Data["datetime"]; ok {
		if t, ok := v.(object.String); ok {
			e.Datetime = string(t)
		}
	}
	return nil
}