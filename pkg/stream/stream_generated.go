// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package stream

import (
	json "encoding/json"

	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

type (
	Policy struct {
		Subjects   []string `json:"subjects:as,omitempty"`
		Resources  []string `json:"resources:as,omitempty"`
		Conditions []string `json:"conditions:as,omitempty"`
		Action     string   `json:"action:s,omitempty"`
	}
	Created struct {
		Nonce           string            `json:"nonce:s,omitempty"`
		CreatedDateTime string            `json:"createdDateTime:s,omitempty"`
		Policies        []*Policy         `json:"policies:ao,omitempty"`
		Signature       *crypto.Signature `json:"@signature:o,omitempty"`
		Identity        crypto.PublicKey  `json:"@identity:s,omitempty"`
	}
	Subscribed struct {
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:s,omitempty"`
	}
	Unsubscribed struct {
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:s,omitempty"`
	}
	PolicyAttached struct {
		Stream    object.Hash       `json:"stream:s,omitempty"`
		Parents   []object.Hash     `json:"parents:as,omitempty"`
		Policies  []*Policy         `json:"policies:ao,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:s,omitempty"`
	}
	RequestEventList struct {
		Stream    object.Hash       `json:"stream:s,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:s,omitempty"`
	}
	EventListCreated struct {
		Stream    object.Hash       `json:"stream:s,omitempty"`
		Events    []object.Hash     `json:"events:as,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:s,omitempty"`
	}
)

func (e *Policy) GetType() string {
	return "nimona.io/stream.Policy"
}

func (e *Policy) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/stream.Policy"
	if len(e.Subjects) > 0 {
		m["subjects:as"] = e.Subjects
	}
	if len(e.Resources) > 0 {
		m["resources:as"] = e.Resources
	}
	if len(e.Conditions) > 0 {
		m["conditions:as"] = e.Conditions
	}
	m["action:s"] = e.Action
	return object.Object(m)
}

func (e *Policy) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Created) GetType() string {
	return "nimona.io/stream.Created"
}

func (e *Created) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/stream.Created"
	m["nonce:s"] = e.Nonce
	m["createdDateTime:s"] = e.CreatedDateTime
	if len(e.Policies) > 0 {
		m["policies:ao"] = func() []interface{} {
			a := make([]interface{}, len(e.Policies))
			for i, v := range e.Policies {
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

func (e *Created) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Subscribed) GetType() string {
	return "nimona.io/stream.Subscribed"
}

func (e *Subscribed) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/stream.Subscribed"
	if e.Signature != nil {
		m["@signature:o"] = e.Signature.ToObject().ToMap()
	}
	m["@identity:s"] = e.Identity
	return object.Object(m)
}

func (e *Subscribed) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Unsubscribed) GetType() string {
	return "nimona.io/stream.Unsubscribed"
}

func (e *Unsubscribed) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/stream.Unsubscribed"
	if e.Signature != nil {
		m["@signature:o"] = e.Signature.ToObject().ToMap()
	}
	m["@identity:s"] = e.Identity
	return object.Object(m)
}

func (e *Unsubscribed) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *PolicyAttached) GetType() string {
	return "nimona.io/stream.PolicyAttached"
}

func (e *PolicyAttached) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/stream.PolicyAttached"
	m["stream:s"] = e.Stream
	if len(e.Parents) > 0 {
		m["parents:as"] = e.Parents
	}
	if len(e.Policies) > 0 {
		m["policies:ao"] = func() []interface{} {
			a := make([]interface{}, len(e.Policies))
			for i, v := range e.Policies {
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

func (e *PolicyAttached) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *RequestEventList) GetType() string {
	return "nimona.io/stream.RequestEventList"
}

func (e *RequestEventList) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/stream.RequestEventList"
	m["stream:s"] = e.Stream
	if e.Signature != nil {
		m["@signature:o"] = e.Signature.ToObject().ToMap()
	}
	m["@identity:s"] = e.Identity
	return object.Object(m)
}

func (e *RequestEventList) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *EventListCreated) GetType() string {
	return "nimona.io/stream.EventListCreated"
}

func (e *EventListCreated) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "nimona.io/stream.EventListCreated"
	m["stream:s"] = e.Stream
	if len(e.Events) > 0 {
		m["events:as"] = e.Events
	}
	if e.Signature != nil {
		m["@signature:o"] = e.Signature.ToObject().ToMap()
	}
	m["@identity:s"] = e.Identity
	return object.Object(m)
}

func (e *EventListCreated) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
