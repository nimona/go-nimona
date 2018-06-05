package wire

import (
	"encoding/json"
)

// Message for exchanging data via wire
type Message struct {
	Extension   string `json:"extension,omitempty"`
	PayloadType string `json:"payload_type,omitempty"`
	Codec       string `json:"codec,omitempty"`
	Payload     []byte `json:"payload,omitempty"`
}

// DecodePayload decodes the message's payload according to the coded,
// and stores the result in the value pointed to by r.
func (h *Message) DecodePayload(r interface{}) error {
	return json.Unmarshal(h.Payload, r)
}

// EncodePayload encodes the given value using the message's codec, and stores
// the result in the message's payload.
func (h *Message) EncodePayload(r interface{}) error {
	bs, err := json.Marshal(r)
	h.Payload = bs
	return err
}
