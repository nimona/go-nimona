// Code generated by nimona.io/tools/objectify. DO NOT EDIT.

// +build !generate

package peer

import (
	"encoding/json"

	"nimona.io/pkg/object"
)

const (
	PeerType = "nimona.io/discovery/peer"
)

// ToObject returns a f12n object
func (s Peer) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s": PeerType,
	}
	b, _ := json.Marshal(s)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

// FromObject populates the struct from a f12n object
func (s *Peer) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, s)
}

// GetType returns the object's type
func (s Peer) GetType() string {
	return PeerType
}
