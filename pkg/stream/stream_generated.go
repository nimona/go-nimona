// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package stream

import (
	json "encoding/json"

	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

type (
	Policy struct {
		Subjects   []*crypto.PublicKey `json:"subjects:ao,omitempty"`
		Resources  []string            `json:"resources:as,omitempty"`
		Conditions []string            `json:"conditions:as,omitempty"`
		Action     string              `json:"action:s,omitempty"`
	}
	Created struct {
		Nonce           string              `json:"nonce:s,omitempty"`
		CreatedDateTime string              `json:"createdDateTime:s,omitempty"`
		Policies        []*Policy           `json:"@policies:ao,omitempty"`
		Signature       *crypto.Signature   `json:"@signature:o,omitempty"`
		Authors         []*crypto.PublicKey `json:"@authors:ao,omitempty"`
	}
	Subscribed struct {
		Policies  []*Policy           `json:"@policies:ao,omitempty"`
		Signature *crypto.Signature   `json:"@signature:o,omitempty"`
		Authors   []*crypto.PublicKey `json:"@authors:ao,omitempty"`
	}
	PolicyAttached struct {
		Stream    *object.Hash        `json:"@stream:o,omitempty"`
		Parents   []*object.Hash      `json:"@parents:ao,omitempty"`
		Policies  []*Policy           `json:"@policies:ao,omitempty"`
		Signature *crypto.Signature   `json:"@signature:o,omitempty"`
		Authors   []*crypto.PublicKey `json:"@authors:ao,omitempty"`
	}
	RequestEventList struct {
		Streams []*object.Hash `json:"streams:ao,omitempty"`
	}
	EventListCreated struct {
		Stream *object.Hash   `json:"stream:o,omitempty"`
		Events []*object.Hash `json:"events:ao,omitempty"`
	}
	RequestEvents struct {
		Events []*object.Hash `json:"events:ao,omitempty"`
	}
)

func (e *Policy) GetType() string {
	return "nimona.io/stream.Policy"
}

func (e *Policy) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/stream.Policy",
		"@domain:s": "nimona.io/stream",
		"@struct:s": "Policy",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *Policy) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Created) EventName() string {
	return "Created"
}

func (e *Created) GetType() string {
	return "nimona.io/stream.Created"
}

func (e *Created) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/stream.Created",
		"@domain:s": "nimona.io/stream",
		"@event:s":  "Created",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *Created) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Subscribed) EventName() string {
	return "Subscribed"
}

func (e *Subscribed) GetType() string {
	return "nimona.io/stream.Subscribed"
}

func (e *Subscribed) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/stream.Subscribed",
		"@domain:s": "nimona.io/stream",
		"@event:s":  "Subscribed",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *Subscribed) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *PolicyAttached) EventName() string {
	return "PolicyAttached"
}

func (e *PolicyAttached) GetType() string {
	return "nimona.io/stream.PolicyAttached"
}

func (e *PolicyAttached) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/stream.PolicyAttached",
		"@domain:s": "nimona.io/stream",
		"@event:s":  "PolicyAttached",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *PolicyAttached) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *RequestEventList) EventName() string {
	return "RequestEventList"
}

func (e *RequestEventList) GetType() string {
	return "nimona.io/stream.RequestEventList"
}

func (e *RequestEventList) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/stream.RequestEventList",
		"@domain:s": "nimona.io/stream",
		"@event:s":  "RequestEventList",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *RequestEventList) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *EventListCreated) EventName() string {
	return "EventListCreated"
}

func (e *EventListCreated) GetType() string {
	return "nimona.io/stream.EventListCreated"
}

func (e *EventListCreated) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/stream.EventListCreated",
		"@domain:s": "nimona.io/stream",
		"@event:s":  "EventListCreated",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *EventListCreated) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *RequestEvents) EventName() string {
	return "RequestEvents"
}

func (e *RequestEvents) GetType() string {
	return "nimona.io/stream.RequestEvents"
}

func (e *RequestEvents) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/stream.RequestEvents",
		"@domain:s": "nimona.io/stream",
		"@event:s":  "RequestEvents",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *RequestEvents) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
