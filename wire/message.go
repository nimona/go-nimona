package wire

import (
	"encoding/json"

	"github.com/coreos/go-semver/semver"
)

type messageOut struct {
	Version     semver.Version `json:"version,omitempty"`
	Codec       string         `json:"codec,omitempty"`
	Extension   string         `json:"extension,omitempty"`
	PayloadType string         `json:"payload_type,omitempty"`
	Payload     interface{}    `json:"payload,omitempty"`
	From        string         `json:"from,omitempty"`
	To          string         `json:"to,omitempty"`
}

type Message struct {
	Version     semver.Version  `json:"version,omitempty"`
	Codec       string          `json:"codec,omitempty"`
	Extension   string          `json:"extension,omitempty"`
	PayloadType string          `json:"payload_type,omitempty"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	From        string          `json:"from,omitempty"`
	To          string          `json:"to,omitempty"`
}

func (h *Message) DecodePayload(r interface{}) error {
	return json.Unmarshal(h.Payload, r)
}

func (m Message) String() string {
	type prettyMessage struct {
		Version     string          `json:"version,omitempty"`
		Codec       string          `json:"codec,omitempty"`
		Extension   string          `json:"extension,omitempty"`
		PayloadType string          `json:"payload_type,omitempty"`
		Payload     json.RawMessage `json:"payload,omitempty"`
		From        string          `json:"from,omitempty"`
		To          string          `json:"to,omitempty"`
	}

	pm := &prettyMessage{
		Version:     m.Version.String(),
		From:        m.From,
		To:          m.To,
		Extension:   m.Extension,
		PayloadType: m.PayloadType,
		Codec:       m.Codec,
		Payload:     m.Payload,
	}

	b, _ := json.MarshalIndent(pm, "", "\t")
	return string(b)
}
