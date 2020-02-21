// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package stream

import (
	"errors"

	crypto "nimona.io/pkg/crypto"
	immutable "nimona.io/pkg/immutable"
	object "nimona.io/pkg/object"
)

type (
	Policy struct {
		raw        object.Object
		Stream     object.Hash
		Parents    []object.Hash
		Owners     []crypto.PublicKey
		Policy     object.Policy
		Signature  object.Signature
		Subjects   []string
		Resources  []string
		Conditions []string
		Action     string
	}
	Request struct {
		raw       object.Object
		Stream    object.Hash
		Parents   []object.Hash
		Owners    []crypto.PublicKey
		Policy    object.Policy
		Signature object.Signature
		Nonce     string
		Leaves    []object.Hash
	}
	Response struct {
		raw       object.Object
		Stream    object.Hash
		Parents   []object.Hash
		Owners    []crypto.PublicKey
		Policy    object.Policy
		Signature object.Signature
		Nonce     string
		Children  []object.Hash
	}
	ObjectRequest struct {
		raw       object.Object
		Stream    object.Hash
		Parents   []object.Hash
		Owners    []crypto.PublicKey
		Policy    object.Policy
		Signature object.Signature
		Nonce     string
		Objects   []object.Hash
	}
	ObjectResponse struct {
		raw       object.Object
		Stream    object.Hash
		Parents   []object.Hash
		Owners    []crypto.PublicKey
		Policy    object.Policy
		Signature object.Signature
		Nonce     string
		Objects   []*object.Object
	}
	Announcement struct {
		raw       object.Object
		Stream    object.Hash
		Parents   []object.Hash
		Owners    []crypto.PublicKey
		Policy    object.Policy
		Signature object.Signature
		Nonce     string
		Leaves    []object.Hash
	}
)

func (e Policy) GetType() string {
	return "nimona.io/stream.Policy"
}

// func (e *Policy) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e Policy) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *Policy) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e Policy) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *Policy) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e Policy) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *Policy) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e Policy) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *Policy) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e Policy) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e Policy) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "subjects",
				Type:       "string",
				Hint:       "s",
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
	o = o.SetType("nimona.io/stream.Policy")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.SetSignature(e.Signature)
	o = o.SetPolicy(e.Policy)
	if len(e.Subjects) > 0 {
		v := immutable.List{}
		for _, iv := range e.Subjects {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("subjects:as", v)
	}
	if len(e.Resources) > 0 {
		v := immutable.List{}
		for _, iv := range e.Resources {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("resources:as", v)
	}
	if len(e.Conditions) > 0 {
		v := immutable.List{}
		for _, iv := range e.Conditions {
			v = v.Append(immutable.String(iv))
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
	data, ok := o.Raw().Value("data:o").(immutable.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signature = o.GetSignature()
	e.Policy = o.GetPolicy()
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

func (e Request) GetType() string {
	return "nimona.io/stream.Request"
}

// func (e *Request) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e Request) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *Request) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e Request) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *Request) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e Request) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *Request) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e Request) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *Request) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e Request) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e Request) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "nonce",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "leaves",
				Type:       "nimona.io/object.Hash",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e Request) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/stream.Request")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.SetSignature(e.Signature)
	o = o.SetPolicy(e.Policy)
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	if len(e.Leaves) > 0 {
		v := immutable.List{}
		for _, iv := range e.Leaves {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("leaves:as", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Request) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(immutable.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signature = o.GetSignature()
	e.Policy = o.GetPolicy()
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("leaves:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Leaves = make([]object.Hash, len(m))
		for i, iv := range m {
			e.Leaves[i] = object.Hash(iv)
		}
	}
	return nil
}

func (e Response) GetType() string {
	return "nimona.io/stream.Response"
}

// func (e *Response) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e Response) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *Response) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e Response) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *Response) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e Response) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *Response) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e Response) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *Response) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e Response) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e Response) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "nonce",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "children",
				Type:       "nimona.io/object.Hash",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e Response) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/stream.Response")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.SetSignature(e.Signature)
	o = o.SetPolicy(e.Policy)
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	if len(e.Children) > 0 {
		v := immutable.List{}
		for _, iv := range e.Children {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("children:as", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Response) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(immutable.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signature = o.GetSignature()
	e.Policy = o.GetPolicy()
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("children:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Children = make([]object.Hash, len(m))
		for i, iv := range m {
			e.Children[i] = object.Hash(iv)
		}
	}
	return nil
}

func (e ObjectRequest) GetType() string {
	return "nimona.io/stream.ObjectRequest"
}

// func (e *ObjectRequest) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e ObjectRequest) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *ObjectRequest) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e ObjectRequest) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *ObjectRequest) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e ObjectRequest) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *ObjectRequest) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e ObjectRequest) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *ObjectRequest) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e ObjectRequest) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e ObjectRequest) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "nonce",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "objects",
				Type:       "nimona.io/object.Hash",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e ObjectRequest) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/stream.ObjectRequest")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.SetSignature(e.Signature)
	o = o.SetPolicy(e.Policy)
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	if len(e.Objects) > 0 {
		v := immutable.List{}
		for _, iv := range e.Objects {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("objects:as", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *ObjectRequest) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(immutable.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signature = o.GetSignature()
	e.Policy = o.GetPolicy()
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("objects:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Objects = make([]object.Hash, len(m))
		for i, iv := range m {
			e.Objects[i] = object.Hash(iv)
		}
	}
	return nil
}

func (e ObjectResponse) GetType() string {
	return "nimona.io/stream.ObjectResponse"
}

// func (e *ObjectResponse) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e ObjectResponse) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *ObjectResponse) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e ObjectResponse) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *ObjectResponse) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e ObjectResponse) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *ObjectResponse) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e ObjectResponse) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *ObjectResponse) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e ObjectResponse) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e ObjectResponse) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "nonce",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "objects",
				Type:       "nimona.io/object.Object",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e ObjectResponse) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/stream.ObjectResponse")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.SetSignature(e.Signature)
	o = o.SetPolicy(e.Policy)
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	if len(e.Objects) > 0 {
		v := immutable.List{}
		for _, iv := range e.Objects {
			v = v.Append(iv.ToObject().Raw())
		}
		o = o.Set("objects:ao", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *ObjectResponse) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(immutable.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signature = o.GetSignature()
	e.Policy = o.GetPolicy()
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("objects:ao"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]interface{})
		e.Objects = make([]*object.Object, len(m))
		for i, iv := range m {
			eo := object.FromMap(iv.(map[string]interface{}))
			e.Objects[i] = &eo
		}
	}
	return nil
}

func (e Announcement) GetType() string {
	return "nimona.io/stream.Announcement"
}

// func (e *Announcement) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e Announcement) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *Announcement) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e Announcement) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *Announcement) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e Announcement) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *Announcement) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e Announcement) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *Announcement) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e Announcement) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e Announcement) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "nonce",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "leaves",
				Type:       "nimona.io/object.Hash",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e Announcement) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/stream.Announcement")
	if len(e.Stream) > 0 {
		o = o.SetStream(e.Stream)
	}
	if len(e.Parents) > 0 {
		o = o.SetParents(e.Parents)
	}
	if len(e.Owners) > 0 {
		o = o.SetOwners(e.Owners)
	}
	o = o.SetSignature(e.Signature)
	o = o.SetPolicy(e.Policy)
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	if len(e.Leaves) > 0 {
		v := immutable.List{}
		for _, iv := range e.Leaves {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("leaves:as", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Announcement) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:o").(immutable.Map)
	if !ok {
		return errors.New("missing data")
	}
	e.raw = object.Object{}
	e.raw = e.raw.SetType(o.GetType())
	e.Stream = o.GetStream()
	e.Parents = o.GetParents()
	e.Owners = o.GetOwners()
	e.Signature = o.GetSignature()
	e.Policy = o.GetPolicy()
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("leaves:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Leaves = make([]object.Hash, len(m))
		for i, iv := range m {
			e.Leaves[i] = object.Hash(iv)
		}
	}
	return nil
}
