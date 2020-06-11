// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package peer

import (
	"errors"

	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

type (
	Certificate struct {
		raw        object.Object
		Stream     object.Hash
		Parents    []object.Hash
		Owners     []crypto.PublicKey
		Policy     object.Policy
		Signatures []object.Signature
		Nonce      string
		Created    string
		Expires    string
	}
	CertificateRequest struct {
		raw                    object.Object
		Stream                 object.Hash
		Parents                []object.Hash
		Owners                 []crypto.PublicKey
		Policy                 object.Policy
		Signatures             []object.Signature
		ApplicationName        string
		ApplicationDescription string
		ApplicationURL         string
		Subject                string
		Resources              []string
		Actions                []string
		Nonce                  string
	}
)

func (e Certificate) GetType() string {
	return "nimona.io/peer.Certificate"
}

func (e Certificate) GetSchema() *object.SchemaObject {
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
				Name:       "created",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "expires",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e Certificate) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/peer.Certificate")
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
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	if e.Created != "" {
		o = o.Set("created:s", e.Created)
	}
	if e.Expires != "" {
		o = o.Set("expires:s", e.Expires)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *Certificate) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:m").(object.Map)
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
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("created:s"); v != nil {
		e.Created = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("expires:s"); v != nil {
		e.Expires = string(v.PrimitiveHinted().(string))
	}
	return nil
}

func (e CertificateRequest) GetType() string {
	return "nimona.io/peer.CertificateRequest"
}

func (e CertificateRequest) GetSchema() *object.SchemaObject {
	return &object.SchemaObject{
		Properties: []*object.SchemaProperty{
			&object.SchemaProperty{
				Name:       "applicationName",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "applicationDescription",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "applicationURL",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "subject",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
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
				Name:       "actions",
				Type:       "string",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&object.SchemaProperty{
				Name:       "nonce",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e CertificateRequest) ToObject() object.Object {
	o := object.Object{}
	o = o.SetType("nimona.io/peer.CertificateRequest")
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
	if e.ApplicationName != "" {
		o = o.Set("applicationName:s", e.ApplicationName)
	}
	if e.ApplicationDescription != "" {
		o = o.Set("applicationDescription:s", e.ApplicationDescription)
	}
	if e.ApplicationURL != "" {
		o = o.Set("applicationURL:s", e.ApplicationURL)
	}
	if e.Subject != "" {
		o = o.Set("subject:s", e.Subject)
	}
	if len(e.Resources) > 0 {
		v := object.List{}
		for _, iv := range e.Resources {
			v = v.Append(object.String(iv))
		}
		o = o.Set("resources:as", v)
	}
	if len(e.Actions) > 0 {
		v := object.List{}
		for _, iv := range e.Actions {
			v = v.Append(object.String(iv))
		}
		o = o.Set("actions:as", v)
	}
	if e.Nonce != "" {
		o = o.Set("nonce:s", e.Nonce)
	}
	// if schema := e.GetSchema(); schema != nil {
	// 	m["_schema:m"] = schema.ToObject().ToMap()
	// }
	return o
}

func (e *CertificateRequest) FromObject(o object.Object) error {
	data, ok := o.Raw().Value("data:m").(object.Map)
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
	if v := data.Value("applicationName:s"); v != nil {
		e.ApplicationName = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("applicationDescription:s"); v != nil {
		e.ApplicationDescription = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("applicationURL:s"); v != nil {
		e.ApplicationURL = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("subject:s"); v != nil {
		e.Subject = string(v.PrimitiveHinted().(string))
	}
	if v := data.Value("resources:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Resources = make([]string, len(m))
		for i, iv := range m {
			e.Resources[i] = string(iv)
		}
	}
	if v := data.Value("actions:as"); v != nil && v.IsList() {
		m := v.PrimitiveHinted().([]string)
		e.Actions = make([]string, len(m))
		for i, iv := range m {
			e.Actions[i] = string(iv)
		}
	}
	if v := data.Value("nonce:s"); v != nil {
		e.Nonce = string(v.PrimitiveHinted().(string))
	}
	return nil
}
