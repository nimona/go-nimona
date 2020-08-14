// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package feed

import (
	"errors"

	object "nimona.io/pkg/object"
)

type (
	FeedStreamRoot struct {
		raw      object.Object
		Metadata object.Metadata
		Type     string
		Datetime string
	}
	Added struct {
		raw        object.Object
		Metadata   object.Metadata
		ObjectHash []object.Hash
		Sequence   int64
		Datetime   string
	}
	Removed struct {
		raw        object.Object
		Metadata   object.Metadata
		ObjectHash []object.Hash
		Sequence   int64
		Datetime   string
	}
)

func (e FeedStreamRoot) GetType() string {
	return "stream:nimona.io/feed"
}

func (e FeedStreamRoot) IsStreamRoot() bool {
	return true
}

func (e FeedStreamRoot) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{{
			Name:       "type",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}, {
			Name:       "datetime",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}},
	}
}

func (e FeedStreamRoot) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("stream:nimona.io/feed")
	if len(e.Metadata.Stream) > 0 {
		o = o.SetStream(e.Metadata.Stream)
	}
	if len(e.Metadata.Parents) > 0 {
		o = o.SetParents(e.Metadata.Parents)
	}
	if !e.Metadata.Owner.IsEmpty() {
		o = o.SetOwner(e.Metadata.Owner)
	}
	if !e.Metadata.Signature.IsEmpty() {
		o = o.SetSignature(e.Metadata.Signature)
	}
	o = o.SetPolicy(e.Metadata.Policy)
	if e.Type != "" {
		o = o.Set("type:s", e.Type)
	}
	if e.Datetime != "" {
		o = o.Set("datetime:s", e.Datetime)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *FeedStreamRoot) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:m").(object.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Metadata.Stream = o.GetStream()
	e.Metadata.Parents = o.GetParents()
	e.Metadata.Owner = o.GetOwner()
	e.Metadata.Signature = o.GetSignature()
	e.Metadata.Policy = o.GetPolicy()
	if v := data.Value("type:s"); v != nil {
		e.Type = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("datetime:s"); v != nil {
		e.Datetime = string(v.PrimitiveHinted().(string))
	}
	return nil
}

func (e Added) GetType() string {
	return "event:nimona.io/feed.Added"
}

func (e Added) IsStreamRoot() bool {
	return false
}

func (e Added) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{{
			Name:       "objectHash",
			Type:       "nimona.io/object.Hash",
			Hint:       "s",
			IsRepeated: true,
			IsOptional: false,
		}, {
			Name:       "sequence",
			Type:       "int",
			Hint:       "i",
			IsRepeated: false,
			IsOptional: false,
		}, {
			Name:       "datetime",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}},
	}
}

func (e Added) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("event:nimona.io/feed.Added")
	if len(e.Metadata.Stream) > 0 {
		o = o.SetStream(e.Metadata.Stream)
	}
	if len(e.Metadata.Parents) > 0 {
		o = o.SetParents(e.Metadata.Parents)
	}
	if !e.Metadata.Owner.IsEmpty() {
		o = o.SetOwner(e.Metadata.Owner)
	}
	if !e.Metadata.Signature.IsEmpty() {
		o = o.SetSignature(e.Metadata.Signature)
	}
	o = o.SetPolicy(e.Metadata.Policy)
	if len(e.ObjectHash) > 0 {
		v := object.List{}
		for _, iv := range e.ObjectHash {
			v = v.Append(object.String(iv))
		}
		o = o.Set("objectHash:as", v)
	}
	o = o.Set("sequence:i", e.Sequence)
	if e.Datetime != "" {
		o = o.Set("datetime:s", e.Datetime)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Added) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:m").(object.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Metadata.Stream = o.GetStream()
	e.Metadata.Parents = o.GetParents()
	e.Metadata.Owner = o.GetOwner()
	e.Metadata.Signature = o.GetSignature()
	e.Metadata.Policy = o.GetPolicy()
	if v := data.Value("objectHash:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.ObjectHash = make([]object.Hash, len(m))
		for i, iv := range m {
			e.ObjectHash[i] = object.Hash(iv)
		}
	}
	if v := data.Value("sequence:i"); v != nil {
		e.Sequence = int64(v.PrimitiveHinted().(int64))
	}
	if v := data.Value("datetime:s"); v != nil {
		e.Datetime = string(v.PrimitiveHinted().(string))
	}
	return nil
}

func (e Removed) GetType() string {
	return "event:nimona.io/feed.Removed"
}

func (e Removed) IsStreamRoot() bool {
	return false
}

func (e Removed) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{{
			Name:       "objectHash",
			Type:       "nimona.io/object.Hash",
			Hint:       "s",
			IsRepeated: true,
			IsOptional: false,
		}, {
			Name:       "sequence",
			Type:       "int",
			Hint:       "i",
			IsRepeated: false,
			IsOptional: false,
		}, {
			Name:       "datetime",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}},
	}
}

func (e Removed) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("event:nimona.io/feed.Removed")
	if len(e.Metadata.Stream) > 0 {
		o = o.SetStream(e.Metadata.Stream)
	}
	if len(e.Metadata.Parents) > 0 {
		o = o.SetParents(e.Metadata.Parents)
	}
	if !e.Metadata.Owner.IsEmpty() {
		o = o.SetOwner(e.Metadata.Owner)
	}
	if !e.Metadata.Signature.IsEmpty() {
		o = o.SetSignature(e.Metadata.Signature)
	}
	o = o.SetPolicy(e.Metadata.Policy)
	if len(e.ObjectHash) > 0 {
		v := object.List{}
		for _, iv := range e.ObjectHash {
			v = v.Append(object.String(iv))
		}
		o = o.Set("objectHash:as", v)
	}
	o = o.Set("sequence:i", e.Sequence)
	if e.Datetime != "" {
		o = o.Set("datetime:s", e.Datetime)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Removed) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:m").(object.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Metadata.Stream = o.GetStream()
	e.Metadata.Parents = o.GetParents()
	e.Metadata.Owner = o.GetOwner()
	e.Metadata.Signature = o.GetSignature()
	e.Metadata.Policy = o.GetPolicy()
	if v := data.Value("objectHash:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.ObjectHash = make([]object.Hash, len(m))
		for i, iv := range m {
			e.ObjectHash[i] = object.Hash(iv)
		}
	}
	if v := data.Value("sequence:i"); v != nil {
		e.Sequence = int64(v.PrimitiveHinted().(int64))
	}
	if v := data.Value("datetime:s"); v != nil {
		e.Datetime = string(v.PrimitiveHinted().(string))
	}
	return nil
}
