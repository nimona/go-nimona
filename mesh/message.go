package mesh

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Message for our wire protocol
type Message struct {
	Version   int
	Nonce     string
	Sender    string
	Recipient string
	Topic     string
	Codec     string
	Payload   []byte
	Checksum  []byte
	Signature []byte
}

func (m Message) String() string {
	type prettyMessage struct {
		Version   int         `json:"version,omitempty"`
		Sender    string      `json:"sender,omitempty"`
		Recipient string      `json:"recipient,omitempty"`
		Topic     string      `json:"topic,omitempty"`
		Codec     string      `json:"codec,omitempty"`
		Payload   interface{} `json:"payload,omitempty"`
		Checksum  string      `json:"checksum,omitempty"`
		Signature string      `json:"signature,omitempty"`
	}

	pm := &prettyMessage{
		Version:   m.Version,
		Sender:    m.Sender,
		Recipient: m.Recipient,
		Topic:     m.Topic,
		Codec:     m.Codec,
		Payload:   string(m.Payload),
		Checksum:  fmt.Sprintf("%x", m.Checksum),
		Signature: fmt.Sprintf("%x", m.Signature),
	}

	if strings.Contains(pm.Codec, "json") {
		pl := map[string]interface{}{}
		json.Unmarshal(m.Payload, &pl)
		pm.Payload = pl
	}

	b, _ := json.MarshalIndent(pm, "", "\t")
	return string(b)
}
