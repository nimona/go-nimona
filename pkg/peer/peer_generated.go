// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package peer

import (
	json "encoding/json"

	crypto "nimona.io/pkg/crypto"
	object "nimona.io/pkg/object"
)

// basic structs
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
	return "nimona.io/peer/Peer"
}

func (e *Peer) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/peer/Peer",
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

// domain events
type (
	PeerRequested struct {
		Keys      []string          `json:"keys:as"`
		Signature *crypto.Signature `json:"@signature:o"`
	}
	PeerUpdated struct {
		Addresses []string          `json:"addresses:as"`
		Signature *crypto.Signature `json:"@signature:o"`
	}
)

func (e *PeerRequested) ContextName() string {
	return "nimona.io/peer/Peer"
}

func (e *PeerRequested) EventName() string {
	return "Requested"
}

func (e *PeerRequested) GetType() string {
	return "nimona.io/peer/Peer.Requested"
}

func (e *PeerRequested) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/peer/Peer.Requested",
		"@domain:s": "nimona.io/peer/Peer",
		"@event:s":  "Requested",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *PeerRequested) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *PeerUpdated) ContextName() string {
	return "nimona.io/peer/Peer"
}

func (e *PeerUpdated) EventName() string {
	return "Updated"
}

func (e *PeerUpdated) GetType() string {
	return "nimona.io/peer/Peer.Updated"
}

func (e *PeerUpdated) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "nimona.io/peer/Peer.Updated",
		"@domain:s": "nimona.io/peer/Peer",
		"@event:s":  "Updated",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *PeerUpdated) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
