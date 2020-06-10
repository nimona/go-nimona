// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package exchange

import (
	crypto "nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	object "nimona.io/pkg/object"
)

type (
	ObjectRequest struct {
		raw        object.Object
		Stream     object.Hash
		Parents    []object.Hash
		Owners     []crypto.PublicKey
		Policy     object.Policy
		Signatures []object.Signature
		ObjectHash object.Hash
	}
	DataForward struct {
		raw        object.Object
		Stream     object.Hash
		Parents    []object.Hash
		Owners     []crypto.PublicKey
		Policy     object.Policy
		Signatures []object.Signature
		Recipient  crypto.PublicKey
		Ephermeral crypto.PublicKey
		Data       []byte
	}
)

func (e ObjectRequest) GetType() string {
	return "nimona.io/exchange.ObjectRequest"
}

func (e ObjectRequest) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "objectHash",
				Type:       "nimona.io/object.Hash",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e ObjectRequest) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/exchange.ObjectRequest")
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
	if e.ObjectHash != "" {
		o = o.Set("objectHash:s", e.ObjectHash)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *ObjectRequest) FromObject(o object.Object) error {
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
	if v := data.Value("objectHash:s"); v != nil {
		e.ObjectHash = object.Hash(v.PrimitiveHinted().(string))
	}
	return nil
}

func (e DataForward) GetType() string {
	return "nimona.io/exchange.DataForward"
}

func (e DataForward) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "recipient",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "ephermeral",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "data",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e DataForward) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/exchange.DataForward")
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
	if e.Recipient != "" {
		o = o.Set("recipient:s", e.Recipient)
	}
	if e.Ephermeral != "" {
		o = o.Set("ephermeral:s", e.Ephermeral)
	}
	if len(e.Data) != 0 {
		o = o.Set("data:d", e.Data)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:o"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *DataForward) FromObject(o object.Object) error {
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
	if v := data.Value("recipient:s"); v != nil {
		e.Recipient = crypto.PublicKey(v.PrimitiveHinted().(string))
	}
	if v := data.Value("ephermeral:s"); v != nil {
		e.Ephermeral = crypto.PublicKey(v.PrimitiveHinted().(string))
	}
	if v := data.Value("data:d"); v != nil {
		e.Data = []byte(v.PrimitiveHinted().([]byte))
	}
	return nil
}
