// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package peer

import (
	"errors"

	crypto "nimona.io/pkg/crypto"
	immutable "nimona.io/pkg/immutable"
	object "nimona.io/pkg/object"
)

type (
	Peer struct {
		raw          object.Object
		Stream       object.Hash
		Parents      []object.Hash
		Owners       []crypto.PublicKey
		Policy       object.Policy
		Signature    object.Signature
		Version      int64
		Addresses    []string
		Bloom        []int64
		ContentTypes []string
		Certificates []*object.Certificate
		Relays       []crypto.PublicKey
	}
	LookupRequest struct {
		raw       object.Object
		Stream    object.Hash
		Parents   []object.Hash
		Owners    []crypto.PublicKey
		Policy    object.Policy
		Signature object.Signature
		Nonce     string
		Bloom     []int64
	}
	LookupResponse struct {
		raw       object.Object
		Stream    object.Hash
		Parents   []object.Hash
		Owners    []crypto.PublicKey
		Policy    object.Policy
		Signature object.Signature
		Nonce     string
		Bloom     []int64
		Peers     []*Peer
	}
)

func (e Peer) GetType() string {
	return "nimona.io/peer.Peer"
}

// func (e *Peer) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e Peer) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *Peer) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e Peer) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *Peer) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e Peer) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *Peer) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e Peer) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *Peer) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e Peer) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e Peer) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "version",
				Type:       "int",
				Hint:       "i",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "addresses",
				Type:       "string",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "bloom",
				Type:       "int",
				Hint:       "i",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "contentTypes",
				Type:       "string",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "certificates",
				Type:       "nimona.io/object.Certificate",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "relays",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e Peer) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/peer.Peer")
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
	o = o.Set("version:i", e.Version)
	if len(e.Addresses) > 0 {
		v := immutable.List{}
		for _, iv := range e.Addresses {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("addresses:as", v)
	}
	if len(e.Bloom) > 0 {
		v := immutable.List{}
		for _, iv := range e.Bloom {
			v = v.Append(immutable.Int(iv))
		}
		o = o.Set("bloom:ai", v)
	}
	if len(e.ContentTypes) > 0 {
		v := immutable.List{}
		for _, iv := range e.ContentTypes {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("contentTypes:as", v)
	}
	if len(e.Certificates) > 0 {
		v := immutable.List{}
		for _, iv := range e.Certificates {
			v = v.Append(iv.ToObject().Raw())
		}
		o = o.Set("certificates:ao", v)
	}
	if len(e.Relays) > 0 {
		v := immutable.List{}
		for _, iv := range e.Relays {
			v = v.Append(immutable.String(iv))
		}
		o = o.Set("relays:as", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Peer) FromObject(o object.Object) error {
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
	if v := data.Value("version:i"); v != nil {
		e.Version = int64(v.PrimitiveHinted().(int64))
	}
	if v := data.Value("addresses:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Addresses = make([]string, len(m))
		for i, iv := range m {
			e.Addresses[i] = string(iv)
		}
	}
	if v := data.Value("bloom:ai"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]int64)
		e.Bloom = make([]int64, len(m))
		for i, iv := range m {
			e.Bloom[i] = int64(iv)
		}
	}
	if v := data.Value("contentTypes:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.ContentTypes = make([]string, len(m))
		for i, iv := range m {
			e.ContentTypes[i] = string(iv)
		}
	}
	if v := data.Value("certificates:ao"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]interface{})
		e.Certificates = make([]*object.Certificate, len(m))
		for i, iv := range m {
			es := &object.Certificate{}
			eo := object.FromMap(iv.(map[string]interface{}))
			es.FromObject(eo)
			e.Certificates[i] = es
		}
	}
	if v := data.Value("relays:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Relays = make([]crypto.PublicKey, len(m))
		for i, iv := range m {
			e.Relays[i] = crypto.PublicKey(iv)
		}
	}
	return nil
}

func (e LookupRequest) GetType() string {
	return "nimona.io/LookupRequest"
}

// func (e *LookupRequest) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e LookupRequest) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *LookupRequest) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e LookupRequest) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *LookupRequest) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e LookupRequest) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *LookupRequest) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e LookupRequest) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *LookupRequest) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e LookupRequest) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e LookupRequest) GetSchema() *object.SchemaObject {
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
				Name:       "bloom",
				Type:       "int",
				Hint:       "i",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e LookupRequest) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/LookupRequest")
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
	if len(e.Bloom) > 0 {
		v := immutable.List{}
		for _, iv := range e.Bloom {
			v = v.Append(immutable.Int(iv))
		}
		o = o.Set("bloom:ai", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *LookupRequest) FromObject(o object.Object) error {
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
	if v := data.Value("bloom:ai"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]int64)
		e.Bloom = make([]int64, len(m))
		for i, iv := range m {
			e.Bloom[i] = int64(iv)
		}
	}
	return nil
}

func (e LookupResponse) GetType() string {
	return "nimona.io/LookupResponse"
}

// func (e *LookupResponse) SetStream(v object.Hash) {
// 	e.raw = e.raw.SetStream(v)
// }

// func (e LookupResponse) GetStream() object.Hash {
// 	return e.raw.GetStream()
// }

// func (e *LookupResponse) SetParents(hashes []object.Hash) {
// 	e.raw = e.raw.SetParents(hashes)
// }

// func (e LookupResponse) GetParents() []object.Hash {
// 	return e.raw.GetParents()
// }

// func (e *LookupResponse) SetPolicy(policy object.Policy) {
// 	e.raw = e.raw.SetPolicy(policy)
// }

// func (e LookupResponse) GetPolicy() object.Policy {
// 	return e.raw.GetPolicy()
// }

// func (e *LookupResponse) SetSignature(v object.Signature) {
// 	e.raw = e.raw.SetSignature(v)
// }

// func (e LookupResponse) GetSignature() object.Signature {
// 	return e.raw.GetSignature()
// }

// func (e *LookupResponse) SetOwners(owners []crypto.PublicKey) {
// 	e.raw = e.raw.SetOwners(owners)
// }

// func (e LookupResponse) GetOwners() []crypto.PublicKey {
// 	return e.raw.GetOwners()
// }

func (e LookupResponse) GetSchema() *object.SchemaObject {
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
				Name:       "bloom",
				Type:       "int",
				Hint:       "i",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "peers",
				Type:       "nimona.io/peer.Peer",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e LookupResponse) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/LookupResponse")
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
	if len(e.Bloom) > 0 {
		v := immutable.List{}
		for _, iv := range e.Bloom {
			v = v.Append(immutable.Int(iv))
		}
		o = o.Set("bloom:ai", v)
	}
	if len(e.Peers) > 0 {
		v := immutable.List{}
		for _, iv := range e.Peers {
			v = v.Append(iv.ToObject().Raw())
		}
		o = o.Set("peers:ao", v)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *LookupResponse) FromObject(o object.Object) error {
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
	if v := data.Value("bloom:ai"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]int64)
		e.Bloom = make([]int64, len(m))
		for i, iv := range m {
			e.Bloom[i] = int64(iv)
		}
	}
	if v := data.Value("peers:ao"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]interface{})
		e.Peers = make([]*Peer, len(m))
		for i, iv := range m {
			es := &Peer{}
			eo := object.FromMap(iv.(map[string]interface{}))
			es.FromObject(eo)
			e.Peers[i] = es
		}
	}
	return nil
}
