// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package peer

import (
	json "encoding/json"

	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
	schema "nimona.io/pkg/schema"
)

type (
	Peer struct {
		Version      int64                 `json:"version:i,omitempty"`
		Addresses    []string              `json:"addresses:as,omitempty"`
		Bloom        []int64               `json:"bloom:ai,omitempty"`
		ContentTypes []string              `json:"contentTypes:as,omitempty"`
		Certificates []*crypto.Certificate `json:"certificates:ao,omitempty"`
		Relays       []crypto.PublicKey    `json:"relays:as,omitempty"`
		Signature    *crypto.Signature     `json:"_signature:o,omitempty"`
		Owners       []crypto.PublicKey    `json:"@owners:as,omitempty"`
	}
	LookupRequest struct {
		Nonce     string             `json:"nonce:s,omitempty"`
		Bloom     []int64            `json:"bloom:ai,omitempty"`
		Signature *crypto.Signature  `json:"_signature:o,omitempty"`
		Owners    []crypto.PublicKey `json:"@owners:as,omitempty"`
	}
	LookupResponse struct {
		Nonce     string             `json:"nonce:s,omitempty"`
		Bloom     []int64            `json:"bloom:ai,omitempty"`
		Peers     []*Peer            `json:"peers:ao,omitempty"`
		Signature *crypto.Signature  `json:"_signature:o,omitempty"`
		Owners    []crypto.PublicKey `json:"@owners:as,omitempty"`
	}
)

func (e Peer) GetType() string {
	return "nimona.io/peer.Peer"
}

func (e Peer) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "version",
				Type:       "int",
				Hint:       "i",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "addresses",
				Type:       "string",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "bloom",
				Type:       "int",
				Hint:       "i",
				IsRepeated: true,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "contentTypes",
				Type:       "string",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "certificates",
				Type:       "nimona.io/crypto.Certificate",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "relays",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "_signature",
				Type:       "nimona.io/crypto.Signature",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "@owners",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e Peer) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/peer.Peer"
	if e.Version != 0 {
		m["version:i"] = e.Version
	}
	if len(e.Addresses) > 0 {
		m["addresses:as"] = e.Addresses
	}
	if len(e.Bloom) > 0 {
		m["bloom:ai"] = e.Bloom
	}
	if len(e.ContentTypes) > 0 {
		m["contentTypes:as"] = e.ContentTypes
	}
	if len(e.Certificates) > 0 {
		m["certificates:ao"] = func() []interface{} {
			a := make([]interface{}, len(e.Certificates))
			for i, v := range e.Certificates {
				a[i] = v.ToObject().ToMap()
			}
			return a
		}()
	}
	if len(e.Relays) > 0 {
		m["relays:as"] = e.Relays
	}
	if e.Signature != nil {
		m["_signature:o"] = e.Signature.ToObject().ToMap()
	}
	if len(e.Owners) > 0 {
		m["@owners:as"] = e.Owners
	}
	if schema := e.GetSchema(); schema != nil {
		m["_schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *Peer) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e LookupRequest) GetType() string {
	return "nimona.io/LookupRequest"
}

func (e LookupRequest) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "nonce",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "bloom",
				Type:       "int",
				Hint:       "i",
				IsRepeated: true,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "_signature",
				Type:       "nimona.io/crypto.Signature",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "@owners",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e LookupRequest) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/LookupRequest"
	if e.Nonce != "" {
		m["nonce:s"] = e.Nonce
	}
	if len(e.Bloom) > 0 {
		m["bloom:ai"] = e.Bloom
	}
	if e.Signature != nil {
		m["_signature:o"] = e.Signature.ToObject().ToMap()
	}
	if len(e.Owners) > 0 {
		m["@owners:as"] = e.Owners
	}
	if schema := e.GetSchema(); schema != nil {
		m["_schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *LookupRequest) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e LookupResponse) GetType() string {
	return "nimona.io/LookupResponse"
}

func (e LookupResponse) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "nonce",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "bloom",
				Type:       "int",
				Hint:       "i",
				IsRepeated: true,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "peers",
				Type:       "nimona.io/peer.Peer",
				Hint:       "o",
				IsRepeated: true,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "_signature",
				Type:       "nimona.io/crypto.Signature",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
			&schema.Property{
				Name:       "@owners",
				Type:       "nimona.io/crypto.PublicKey",
				Hint:       "s",
				IsRepeated: true,
				IsOptional: false,
			},
		},
	}
}

func (e LookupResponse) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/LookupResponse"
	if e.Nonce != "" {
		m["nonce:s"] = e.Nonce
	}
	if len(e.Bloom) > 0 {
		m["bloom:ai"] = e.Bloom
	}
	if len(e.Peers) > 0 {
		m["peers:ao"] = func() []interface{} {
			a := make([]interface{}, len(e.Peers))
			for i, v := range e.Peers {
				a[i] = v.ToObject().ToMap()
			}
			return a
		}()
	}
	if e.Signature != nil {
		m["_signature:o"] = e.Signature.ToObject().ToMap()
	}
	if len(e.Owners) > 0 {
		m["@owners:as"] = e.Owners
	}
	if schema := e.GetSchema(); schema != nil {
		m["_schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *LookupResponse) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
