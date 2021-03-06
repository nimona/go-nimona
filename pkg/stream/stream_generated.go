// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package stream

import (
	object "nimona.io/pkg/object"
)

type (
	Policy struct {
		Metadata   object.Metadata `nimona:"metadata:m,omitempty"`
		Subjects   []string
		Resources  []string
		Conditions []string
		Action     string
	}
	Request struct {
		Metadata  object.Metadata `nimona:"metadata:m,omitempty"`
		RequestID string
		RootCID   object.CID
	}
	Response struct {
		Metadata  object.Metadata `nimona:"metadata:m,omitempty"`
		RequestID string
		RootCID   object.CID
		Leaves    []object.CID
	}
	Announcement struct {
		Metadata   object.Metadata `nimona:"metadata:m,omitempty"`
		StreamCID  object.CID
		ObjectCIDs []object.CID
	}
	Subscription struct {
		Metadata object.Metadata `nimona:"metadata:m,omitempty"`
		RootCIDs []object.CID
		Expiry   string
	}
)

func (e *Policy) Type() string {
	return "nimona.io/stream.Policy"
}

func (e Policy) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/stream.Policy",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	// if $member.IsRepeated
	if len(e.Subjects) > 0 {
		// else
		// r.Data["subjects"] = object.ToStringArray(e.Subjects)
		rv := make(object.StringArray, len(e.Subjects))
		for i, iv := range e.Subjects {
			rv[i] = object.String(iv)
		}
		r.Data["subjects"] = rv
	}
	// if $member.IsRepeated
	if len(e.Resources) > 0 {
		// else
		// r.Data["resources"] = object.ToStringArray(e.Resources)
		rv := make(object.StringArray, len(e.Resources))
		for i, iv := range e.Resources {
			rv[i] = object.String(iv)
		}
		r.Data["resources"] = rv
	}
	// if $member.IsRepeated
	if len(e.Conditions) > 0 {
		// else
		// r.Data["conditions"] = object.ToStringArray(e.Conditions)
		rv := make(object.StringArray, len(e.Conditions))
		for i, iv := range e.Conditions {
			rv[i] = object.String(iv)
		}
		r.Data["conditions"] = rv
	}
	// else
	// r.Data["action"] = object.String(e.Action)
	r.Data["action"] = object.String(e.Action)
	return r
}

func (e *Policy) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["subjects"]; ok {
		if t, ok := v.(object.StringArray); ok {
			rv := make([]string, len(t))
			for i, iv := range t {
				rv[i] = string(iv)
			}
			e.Subjects = rv
		}
	}
	if v, ok := o.Data["resources"]; ok {
		if t, ok := v.(object.StringArray); ok {
			rv := make([]string, len(t))
			for i, iv := range t {
				rv[i] = string(iv)
			}
			e.Resources = rv
		}
	}
	if v, ok := o.Data["conditions"]; ok {
		if t, ok := v.(object.StringArray); ok {
			rv := make([]string, len(t))
			for i, iv := range t {
				rv[i] = string(iv)
			}
			e.Conditions = rv
		}
	}
	if v, ok := o.Data["action"]; ok {
		if t, ok := v.(object.String); ok {
			e.Action = string(t)
		}
	}
	return nil
}

func (e *Request) Type() string {
	return "nimona.io/stream.Request"
}

func (e Request) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/stream.Request",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	// else
	// r.Data["requestID"] = object.String(e.RequestID)
	r.Data["requestID"] = object.String(e.RequestID)
	// else if $member.IsPrimitive
	r.Data["rootCID"] = object.String(e.RootCID)
	return r
}

func (e *Request) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["requestID"]; ok {
		if t, ok := v.(object.String); ok {
			e.RequestID = string(t)
		}
	}
	if v, ok := o.Data["rootCID"]; ok {
		if t, ok := v.(object.String); ok {
			e.RootCID = object.CID(t)
		}
	}
	return nil
}

func (e *Response) Type() string {
	return "nimona.io/stream.Response"
}

func (e Response) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/stream.Response",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	// else
	// r.Data["requestID"] = object.String(e.RequestID)
	r.Data["requestID"] = object.String(e.RequestID)
	// else if $member.IsPrimitive
	r.Data["rootCID"] = object.String(e.RootCID)
	// if $member.IsRepeated
	if len(e.Leaves) > 0 {
		// else
		// r.Data["leaves"] = object.ToStringArray(e.Leaves)
		rv := make(object.StringArray, len(e.Leaves))
		for i, iv := range e.Leaves {
			rv[i] = object.String(iv)
		}
		r.Data["leaves"] = rv
	}
	return r
}

func (e *Response) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["requestID"]; ok {
		if t, ok := v.(object.String); ok {
			e.RequestID = string(t)
		}
	}
	if v, ok := o.Data["rootCID"]; ok {
		if t, ok := v.(object.String); ok {
			e.RootCID = object.CID(t)
		}
	}
	if v, ok := o.Data["leaves"]; ok {
		if t, ok := v.(object.StringArray); ok {
			rv := make([]object.CID, len(t))
			for i, iv := range t {
				rv[i] = object.CID(iv)
			}
			e.Leaves = rv
		}
	}
	return nil
}

func (e *Announcement) Type() string {
	return "nimona.io/stream.Announcement"
}

func (e Announcement) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/stream.Announcement",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	// else if $member.IsPrimitive
	r.Data["streamCID"] = object.String(e.StreamCID)
	// if $member.IsRepeated
	if len(e.ObjectCIDs) > 0 {
		// else
		// r.Data["objectCIDs"] = object.ToStringArray(e.ObjectCIDs)
		rv := make(object.StringArray, len(e.ObjectCIDs))
		for i, iv := range e.ObjectCIDs {
			rv[i] = object.String(iv)
		}
		r.Data["objectCIDs"] = rv
	}
	return r
}

func (e *Announcement) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["streamCID"]; ok {
		if t, ok := v.(object.String); ok {
			e.StreamCID = object.CID(t)
		}
	}
	if v, ok := o.Data["objectCIDs"]; ok {
		if t, ok := v.(object.StringArray); ok {
			rv := make([]object.CID, len(t))
			for i, iv := range t {
				rv[i] = object.CID(iv)
			}
			e.ObjectCIDs = rv
		}
	}
	return nil
}

func (e *Subscription) Type() string {
	return "nimona.io/stream.Subscription"
}

func (e Subscription) ToObject() *object.Object {
	r := &object.Object{
		Type:     "nimona.io/stream.Subscription",
		Metadata: e.Metadata,
		Data:     object.Map{},
	}
	// if $member.IsRepeated
	if len(e.RootCIDs) > 0 {
		// else
		// r.Data["rootCIDs"] = object.ToStringArray(e.RootCIDs)
		rv := make(object.StringArray, len(e.RootCIDs))
		for i, iv := range e.RootCIDs {
			rv[i] = object.String(iv)
		}
		r.Data["rootCIDs"] = rv
	}
	// else
	// r.Data["expiry"] = object.String(e.Expiry)
	r.Data["expiry"] = object.String(e.Expiry)
	return r
}

func (e *Subscription) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["rootCIDs"]; ok {
		if t, ok := v.(object.StringArray); ok {
			rv := make([]object.CID, len(t))
			for i, iv := range t {
				rv[i] = object.CID(iv)
			}
			e.RootCIDs = rv
		}
	}
	if v, ok := o.Data["expiry"]; ok {
		if t, ok := v.(object.String); ok {
			e.Expiry = string(t)
		}
	}
	return nil
}
