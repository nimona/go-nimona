package wire

import (
	"encoding/json"
)

type messageOut struct {
	Extension   string      `json:"extension,omitempty"`
	PayloadType string      `json:"payload_type,omitempty"`
	Payload     interface{} `json:"payload,omitempty"`
}

type Message struct {
	Extension   string          `json:"extension,omitempty"`
	PayloadType string          `json:"payload_type,omitempty"`
	Payload     json.RawMessage `json:"payload,omitempty"`
}

func (h *Message) DecodePayload(r interface{}) error {
	return json.Unmarshal(h.Payload, r)
}

func (m Message) String() string {
	type prettyMessage struct {
		Extension   string          `json:"extension,omitempty"`
		PayloadType string          `json:"payload_type,omitempty"`
		Payload     json.RawMessage `json:"payload,omitempty"`
	}

	pm := &prettyMessage{
		Extension:   m.Extension,
		PayloadType: m.PayloadType,
		Payload:     m.Payload,
	}

	b, _ := json.MarshalIndent(pm, "", "\t")
	return string(b)
}
