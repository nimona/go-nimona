// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package stream

import (
	"errors"

	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

type (
	Policy struct {
		raw        object.Object
		Stream     object.Hash
		Parents    []object.Hash
		Owners     []crypto.PublicKey
		Policy     object.Policy
		Signatures []object.Signature
		Subjects   []*crypto.PublicKey
		Resources  []string
		Conditions []string
		Action     string
	}
	Created struct {
		raw             object.Object
		Stream          object.Hash
		Parents         []object.Hash
		Owners          []crypto.PublicKey
		Policy          object.Policy
		Signatures      []object.Signature
		CreatedDateTime string
		PartitionKeys   []string
		Policies        []*Policy
	}
	PoliciesUpdated struct {
		raw        object.Object
		Stream     object.Hash
		Parents    []object.Hash
		Owners     []crypto.PublicKey
		Policy     object.Policy
		Signatures []object.Signature
		Stream     *crypto.Hash
		Parents    []*crypto.Hash
		Policies   []*Policy
	}
)

func (e Policy) GetType() string {
	return "example/stream.Policy"
}

func (e Policy) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "subjects",
				Type:       "example/crypto.PublicKey",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "resources",
				Type:       "string",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "conditions",
				Type:       "string",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "action",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e Policy) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("example/stream.Policy")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.AddSignature(e.Signatures...)
	o = o.SetPolicy(e.Policy)
	if len(e.Subjects) > 0 {
		v := object.List{}
		for _, iv := range e.Subjects {
			v = v.Append(iv.ToObject().Raw())
		}
		o = o.Set("subjects:ao", v)
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
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Policy) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(object.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signatures = o.GetSignatures()
	e.Policy = o.GetPolicy()
	if v := data.Value("subjects:ao"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]interface{})
		e.Subjects = make([]*crypto.PublicKey, len(m))
		for i, iv := range m {
			es := &crypto.PublicKey{}
			eo := object.FromMap(iv.(map[string]interface{}))
			es.FromObject(eo)
			e.Subjects[i] = es
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

func (e Created) GetType() string {
	return "example/stream.Created"
}

func (e Created) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "createdDateTime",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "partitionKeys",
				Type:       "string",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "policies",
				Type:       "Policy",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e Created) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("example/stream.Created")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.AddSignature(e.Signatures...)
	o = o.SetPolicy(e.Policy)
	if e.CreatedDateTime != "" {
		o = o.Set("createdDateTime:s", e.CreatedDateTime)
	}
	if len(e.PartitionKeys) > 0 {
		v := object.List{}
		for _, iv := range e.PartitionKeys {
			v = v.Append(object.String(iv))
		}
		o = o.Set("partitionKeys:as", v)
	}
	if len(e.Policies) > 0 {
		v := object.List{}
		for _, iv := range e.Policies {
			v = v.Append(iv.ToObject().Raw())
		}
		o = o.Set("policies:ao", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Created) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(object.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signatures = o.GetSignatures()
	e.Policy = o.GetPolicy()
	if v := data.Value("createdDateTime:s"); v != nil {
		e.CreatedDateTime = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("partitionKeys:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.PartitionKeys = make([]string, len(m))
		for i, iv := range m {
			e.PartitionKeys[i] = string(iv)
		}
	}
	if v := data.Value("policies:ao"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]interface{})
		e.Policies = make([]*Policy, len(m))
		for i, iv := range m {
			es := &Policy{}
			eo := object.FromMap(iv.(map[string]interface{}))
			es.FromObject(eo)
			e.Policies[i] = es
		}
	}
	return nil
}

func (e PoliciesUpdated) GetType() string {
	return "example/stream.PoliciesUpdated"
}

func (e PoliciesUpdated) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "stream",
				Type:       "example/crypto.Hash",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "parents",
				Type:       "example/crypto.Hash",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "policies",
				Type:       "Policy",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e PoliciesUpdated) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("example/stream.PoliciesUpdated")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.AddSignature(e.Signatures...)
	o = o.SetPolicy(e.Policy)
	if e.Stream != nil {
		o = o.Set("stream:o", e.Stream.ToObject().Raw())
	}
	if len(e.Parents) > 0 {
		v := object.List{}
		for _, iv := range e.Parents {
			v = v.Append(iv.ToObject().Raw())
		}
		o = o.Set("parents:ao", v)
	}
	if len(e.Policies) > 0 {
		v := object.List{}
		for _, iv := range e.Policies {
			v = v.Append(iv.ToObject().Raw())
		}
		o = o.Set("policies:ao", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *PoliciesUpdated) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(object.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signatures = o.GetSignatures()
	e.Policy = o.GetPolicy()
	if v := data.Value("stream:o"); v != nil {
		es := &crypto.Hash{}
		eo := object.FromMap(v.PrimitiveHinted().(map[string]interface{}))
		es.FromObject(eo)
		e.Stream = es
	}
	if v := data.Value("parents:ao"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]interface{})
		e.Parents = make([]*crypto.Hash, len(m))
		for i, iv := range m {
			es := &crypto.Hash{}
			eo := object.FromMap(iv.(map[string]interface{}))
			es.FromObject(eo)
			e.Parents[i] = es
		}
	}
	if v := data.Value("policies:ao"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]interface{})
		e.Policies = make([]*Policy, len(m))
		for i, iv := range m {
			es := &Policy{}
			eo := object.FromMap(iv.(map[string]interface{}))
			es.FromObject(eo)
			e.Policies[i] = es
		}
	}
	return nil
}
