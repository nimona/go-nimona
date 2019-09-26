// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package peer

import (
	json "encoding/json"

	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

type (
	Peer struct {
		Addresses []string          `json:"addresses:as"`
		Signature *crypto.Signature `json:"@signature:o"`
	}
)

func (e *Peer) ContextName() string {
	return "nimona.io/peer"
}

func (e *Peer) GetType() string {
	return "Peer"
}

func (e *Peer) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "Peer",
		"@struct:s": "Peer",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *Peer) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

type (
	Requested struct {
		Keys      []string            `json:"keys:as"`
		Signature *crypto.Signature   `json:"@signature:o"`
		Authors   []*crypto.PublicKey `json:"@authors:ao"`
	}
	Updated struct {
		Addresses []string            `json:"addresses:as"`
		Signature *crypto.Signature   `json:"@signature:o"`
		Authors   []*crypto.PublicKey `json:"@authors:ao"`
	}
)

func (e *Requested) EventName() string {
	return "Requested"
}

func (e *Requested) GetType() string {
	return "Peer.Requested"
}

func (e *Requested) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "Peer.Requested",
		"@domain:s": "Peer",
		"@event:s":  "Requested",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *Requested) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *Updated) EventName() string {
	return "Updated"
}

func (e *Updated) GetType() string {
	return "Peer.Updated"
}

func (e *Updated) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "Peer.Updated",
		"@domain:s": "Peer",
		"@event:s":  "Updated",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *Updated) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
