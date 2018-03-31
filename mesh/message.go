package mesh

import (
	"encoding/json"
	"fmt"
	"strings"
)

// type Message interface {
// 	GetVersion() int
// 	GetSender() string
// 	GetRecipient() string
// 	GetTopics() []string
// 	GetPayload() []byte
// 	GetChecksum() []byte
// }

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
		Version   int         `json:"Version,omitempty"`
		Sender    string      `json:"Sender,omitempty"`
		Recipient string      `json:"Recipient,omitempty"`
		Topic     string      `json:"Topic,omitempty"`
		Codec     string      `json:"Codec,omitempty"`
		Payload   interface{} `json:"Payload,omitempty"`
		Checksum  string      `json:"Checksum,omitempty"`
		Signature string      `json:"Signature,omitempty"`
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
