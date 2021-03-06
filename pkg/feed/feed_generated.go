// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package feed

import (
	object "nimona.io/pkg/object"
)

type (
	FeedStreamRoot struct {
		Metadata   object.Metadata `nimona:"metadata:m,omitempty"`
		ObjectType string
		Datetime   string
	}
	Added struct {
		Metadata  object.Metadata `nimona:"metadata:m,omitempty"`
		ObjectCID []object.CID
		Sequence  int64
		Datetime  string
	}
	Removed struct {
		Metadata  object.Metadata `nimona:"metadata:m,omitempty"`
		ObjectCID []object.CID
		Sequence  int64
		Datetime  string
	}
)

func (e *FeedStreamRoot) Type() string {
	return "stream:nimona.io/feed"
}

func (e FeedStreamRoot) ToObject() *object.Object {
	r := &object.Object{
		Type:     "stream:nimona.io/feed",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	// else
	// r.Data["objectType"] = object.String(e.ObjectType)
	r.Data["objectType"] = object.String(e.ObjectType)
	// else
	// r.Data["datetime"] = object.String(e.Datetime)
	r.Data["datetime"] = object.String(e.Datetime)
	return r
}

func (e *FeedStreamRoot) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["objectType"]; ok {
		if t, ok := v.(object.String); ok {
			e.ObjectType = string(t)
		}
	}
	if v, ok := o.Data["datetime"]; ok {
		if t, ok := v.(object.String); ok {
			e.Datetime = string(t)
		}
	}
	return nil
}

func (e *Added) Type() string {
	return "event:nimona.io/feed.Added"
}

func (e Added) ToObject() *object.Object {
	r := &object.Object{
		Type:     "event:nimona.io/feed.Added",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	// if $member.IsRepeated
	if len(e.ObjectCID) > 0 {
		// else
		// r.Data["objectCID"] = object.ToStringArray(e.ObjectCID)
		rv := make(object.StringArray, len(e.ObjectCID))
		for i, iv := range e.ObjectCID {
			rv[i] = object.String(iv)
		}
		r.Data["objectCID"] = rv
	}
	// else
	// r.Data["sequence"] = object.Int(e.Sequence)
	r.Data["sequence"] = object.Int(e.Sequence)
	// else
	// r.Data["datetime"] = object.String(e.Datetime)
	r.Data["datetime"] = object.String(e.Datetime)
	return r
}

func (e *Added) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["objectCID"]; ok {
		if t, ok := v.(object.StringArray); ok {
			rv := make([]object.CID, len(t))
			for i, iv := range t {
				rv[i] = object.CID(iv)
			}
			e.ObjectCID = rv
		}
	}
	if v, ok := o.Data["sequence"]; ok {
		if t, ok := v.(object.Int); ok {
			e.Sequence = int64(t)
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
	return "event:nimona.io/feed.Removed"
}

func (e Removed) ToObject() *object.Object {
	r := &object.Object{
		Type:     "event:nimona.io/feed.Removed",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	// if $member.IsRepeated
	if len(e.ObjectCID) > 0 {
		// else
		// r.Data["objectCID"] = object.ToStringArray(e.ObjectCID)
		rv := make(object.StringArray, len(e.ObjectCID))
		for i, iv := range e.ObjectCID {
			rv[i] = object.String(iv)
		}
		r.Data["objectCID"] = rv
	}
	// else
	// r.Data["sequence"] = object.Int(e.Sequence)
	r.Data["sequence"] = object.Int(e.Sequence)
	// else
	// r.Data["datetime"] = object.String(e.Datetime)
	r.Data["datetime"] = object.String(e.Datetime)
	return r
}

func (e *Removed) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["objectCID"]; ok {
		if t, ok := v.(object.StringArray); ok {
			rv := make([]object.CID, len(t))
			for i, iv := range t {
				rv[i] = object.CID(iv)
			}
			e.ObjectCID = rv
		}
	}
	if v, ok := o.Data["sequence"]; ok {
		if t, ok := v.(object.Int); ok {
			e.Sequence = int64(t)
		}
	}
	if v, ok := o.Data["datetime"]; ok {
		if t, ok := v.(object.String); ok {
			e.Datetime = string(t)
		}
	}
	return nil
}
