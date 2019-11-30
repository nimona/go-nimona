// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package peer

import (
	json "encoding/json"

	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

type (
	Peer struct {
		Addresses    []string              `json:"addresses:as,omitempty"`
		Bloom        []int64               `json:"bloom:ai,omitempty"`
		ContentTypes []string              `json:"contentTypes:as,omitempty"`
		Certificates []*crypto.Certificate `json:"certificates:ao,omitempty"`
		Signature    *crypto.Signature     `json:"@signature:o,omitempty"`
		Identity     crypto.PublicKey      `json:"@identity:s,omitempty"`
	}
	Request struct {
		Bloom     []int64           `json:"bloom:ai,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:s,omitempty"`
	}
	Lookup struct {
		Bloom     []int64           `json:"bloom:ai,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:s,omitempty"`
	}
)

func (e *Peer) GetType() string {
	return "nimona.io/peer.Peer"
}

func (e *Peer) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/peer.Peer"
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
	if e.Signature != nil {
		m["@signature:o"] = e.Signature.ToObject().ToMap()
	}
	m["@identity:s"] = e.Identity
	return object.Object(m)
}

func (e *Peer) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Request) GetType() string {
	return "nimona.io/peer.Request"
}

func (e *Request) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/peer.Request"
	if len(e.Bloom) > 0 {
		m["bloom:ai"] = e.Bloom
	}
	if e.Signature != nil {
		m["@signature:o"] = e.Signature.ToObject().ToMap()
	}
	m["@identity:s"] = e.Identity
	return object.Object(m)
}

func (e *Request) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Lookup) GetType() string {
	return "nimona.io/peer.Lookup"
}

func (e *Lookup) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/peer.Lookup"
	if len(e.Bloom) > 0 {
		m["bloom:ai"] = e.Bloom
	}
	if e.Signature != nil {
		m["@signature:o"] = e.Signature.ToObject().ToMap()
	}
	m["@identity:s"] = e.Identity
	return object.Object(m)
}

func (e *Lookup) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
