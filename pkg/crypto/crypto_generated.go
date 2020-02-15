// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package crypto

import (
	json "encoding/json"

	object "nimona.io/pkg/object"
	schema "nimona.io/pkg/schema"
)

type (
	Certificate struct {
		Subject   PublicKey  `json:"subject:s,omitempty"`
		Created   string     `json:"created:s,omitempty"`
		Expires   string     `json:"expires:s,omitempty"`
		Signature *Signature `json:"@signature:o,omitempty"`
	}
	Signature struct {
		Signer PublicKey `json:"signer:s,omitempty"`
		Alg    string    `json:"alg:s,omitempty"`
		X      []byte    `json:"x:d,omitempty"`
	}
)

func (e Certificate) GetType() string {
	return "nimona.io/crypto.Certificate"
}

func (e Certificate) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "subject",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "created",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "expires",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "@signature",
				Type:       "Signature",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e Certificate) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/crypto.Certificate"
	if e.Subject != "" {
		m["subject:s"] = e.Subject
	}
	if e.Created != "" {
		m["created:s"] = e.Created
	}
	if e.Expires != "" {
		m["expires:s"] = e.Expires
	}
	if e.Signature != nil {
		m["@signature:o"] = e.Signature.ToObject().ToMap()
	}
	if schema := e.GetSchema(); schema != nil {
		m["_schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *Certificate) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e Signature) GetType() string {
	return "nimona.io/crypto.Signature"
}

func (e Signature) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "signer",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "alg",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "x",
				Type:       "data",
				Hint:       "d",
				IsRepeated: false,
				IsOptional: false,
			},
		},
	}
}

func (e Signature) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/crypto.Signature"
	if e.Signer != "" {
		m["signer:s"] = e.Signer
	}
	if e.Alg != "" {
		m["alg:s"] = e.Alg
	}
	if len(e.X) != 0 {
		m["x:d"] = e.X
	}
	if schema := e.GetSchema(); schema != nil {
		m["_schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *Signature) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
