// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package fixtures

import (
	"errors"

	object "nimona.io/pkg/object"
)

type (
	TestPolicy struct {
		raw        object.Object
		Metadata   object.Metadata
		Subjects   []string
		Resources  []string
		Conditions []string
		Action     string
	}
	TestStream struct {
		raw             object.Object
		Metadata        object.Metadata
		Nonce           string
		CreatedDateTime string
	}
	TestSubscribed struct {
		raw      object.Object
		Metadata object.Metadata
		Nonce    string
	}
	TestUnsubscribed struct {
		raw      object.Object
		Metadata object.Metadata
		Nonce    string
	}
)

func (e TestPolicy) GetType() string {
	return "nimona.io/fixtures.TestPolicy"
}

func (e TestPolicy) IsStreamRoot() bool {
	return false
}

func (e TestPolicy) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{{
			Name:       "subjects",
			Type:       "string",
			Hint:       "s",
			IsRepeated: true,
			IsOptional: false,
		}, {
			Name:       "resources",
			Type:       "string",
			Hint:       "s",
			IsRepeated: true,
			IsOptional: false,
		}, {
			Name:       "conditions",
			Type:       "string",
			Hint:       "s",
			IsRepeated: true,
			IsOptional: false,
		}, {
			Name:       "action",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}},
	}
}

func (e TestPolicy) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/fixtures.TestPolicy")
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
	if len(e.Subjects) > 0 {
		v := object.List{}
		for _, iv := range e.Subjects {
			v = v.Append(object.String(iv))
		}
		o = o.Set("subjects:as", v)
	}
	if len(e.Resources) > 0 {
		v := object.List{}
		for _, iv := range e.Resources {
			v = v.Append(object.String(iv))
		}
		o = o.Set("resources:as", v)
	}
	if len(e.Conditions) > 0 {
		v := object.List{}
		for _, iv := range e.Conditions {
			v = v.Append(object.String(iv))
		}
		o = o.Set("conditions:as", v)
	}
	if e.Action != "" {
		o = o.Set("action:s", e.Action)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *TestPolicy) FromObject(o object.Object) error {
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
	if v := data.Value("subjects:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Subjects = make([]string, len(m))
		for i, iv := range m {
			e.Subjects[i] = string(iv)
		}
	}
	if v := data.Value("resources:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Resources = make([]string, len(m))
		for i, iv := range m {
			e.Resources[i] = string(iv)
		}
	}
	if v := data.Value("conditions:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Conditions = make([]string, len(m))
		for i, iv := range m {
			e.Conditions[i] = string(iv)
		}
	}
	if v := data.Value("action:s"); v != nil {
		e.Action = string(v.PrimitiveHinted().(string))
	}
	return nil
}

func (e TestStream) GetType() string {
	return "nimona.io/fixtures.TestStream"
}

func (e TestStream) IsStreamRoot() bool {
	return false
}

func (e TestStream) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{{
			Name:       "nonce",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}, {
			Name:       "createdDateTime",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}},
	}
}

func (e TestStream) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/fixtures.TestStream")
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
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	if e.CreatedDateTime != "" {
		o = o.Set("createdDateTime:s", e.CreatedDateTime)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *TestStream) FromObject(o object.Object) error {
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
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("createdDateTime:s"); v != nil {
		e.CreatedDateTime = string(v.PrimitiveHinted().(string))
	}
	return nil
}

func (e TestSubscribed) GetType() string {
	return "nimona.io/fixtures.TestSubscribed"
}

func (e TestSubscribed) IsStreamRoot() bool {
	return false
}

func (e TestSubscribed) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{{
			Name:       "nonce",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}},
	}
}

func (e TestSubscribed) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/fixtures.TestSubscribed")
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
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *TestSubscribed) FromObject(o object.Object) error {
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
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	return nil
}

func (e TestUnsubscribed) GetType() string {
	return "nimona.io/fixtures.TestUnsubscribed"
}

func (e TestUnsubscribed) IsStreamRoot() bool {
	return false
}

func (e TestUnsubscribed) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{{
			Name:       "nonce",
			Type:       "string",
			Hint:       "s",
			IsRepeated: false,
			IsOptional: false,
		}},
	}
}

func (e TestUnsubscribed) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/fixtures.TestUnsubscribed")
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
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *TestUnsubscribed) FromObject(o object.Object) error {
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
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	return nil
}
