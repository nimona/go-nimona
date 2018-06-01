package wire

import (
	"encoding/json"
)

type Message struct {
	Extension   string `json:"extension,omitempty"`
	PayloadType string `json:"payload_type,omitempty"`
	Codec       string `json:"codec,omitempty"`
	Payload     []byte `json:"payload,omitempty"`
}

func (h *Message) DecodePayload(r interface{}) error {
	return json.Unmarshal(h.Payload, r)
}

func (h *Message) EncodePayload(r interface{}) error {
	bs, err := json.Marshal(r)
	h.Payload = bs
	return err
}
