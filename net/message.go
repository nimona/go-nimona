package net

import (
	"github.com/ugorji/go/codec"
)

func NewMessage(contentType string, recipients []string, payload interface{}) (*Message, error) {
	message := &Message{
		Version: 0,
		Type:    contentType,
		Headers: Headers{
			Recipients: recipients,
		},
		Payload: payload,
	}
	return message, nil
}

type Headers struct {
	Recipients []string `json:"recipients,omitempty"`
	Signer     string   `json:"signer,omitempty"`
}

// Message for exchanging data via the messenger
type Message struct {
	Version   uint        `json:"version"`
	Type      string      `json:"@type"`
	Headers   Headers     `json:"headers,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Signature []byte      `json:"signature,omitempty"`
}

func (message *Message) IsSigned() bool {
	// TODO make this part of the message and digest?
	return message.Signature != nil && len(message.Signature) > 0
}

func getMessageDigest(message *Message) ([]byte, error) {
	digest := []interface{}{
		message.Version,
		message.Type,
		message.Headers,
		message.Payload,
	}

	digestBytes, err := Marshal(digest)
	if err != nil {
		return nil, err
	}

	return digestBytes, nil
}

func (message *Message) Sign(signerPeerInfo *SecretPeerInfo) error {
	message.Headers.Signer = signerPeerInfo.ID
	digest, err := getMessageDigest(message)
	if err != nil {
		return err
	}

	signature, err := Sign(digest, signerPeerInfo.PrivateKey)
	if err != nil {
		return err
	}

	message.Signature = signature
	return nil
}

func (message *Message) Verify() error {
	digest, err := getMessageDigest(message)
	if err != nil {
		return err
	}

	return Verify(message.Headers.Signer, digest, message.Signature)
}

func Marshal(o interface{}) ([]byte, error) {
	b := []byte{}
	enc := codec.NewEncoderBytes(&b, &codec.CborHandle{})
	if err := enc.Encode(o); err != nil {
		return nil, err
	}

	return b, nil
}

func Unmarshal(b []byte) (*Message, error) {
	m := &Message{}
	dec := codec.NewDecoderBytes(b, &codec.CborHandle{})
	if err := dec.Decode(m); err != nil {
		return nil, err
	}
	ct := GetContentType(m.Type)
	if err := m.DecodePayload(&ct); err != nil {
		return nil, err
	}
	m.Payload = ct
	return m, nil
}

// DecodePayload decodes the message's payload according to the coded,
// and stores the result in the value pointed to by r.
func (h *Message) DecodePayload(r interface{}) error {
	enc, err := Marshal(h.Payload)
	if err != nil {
		return err
	}

	dec := codec.NewDecoderBytes(enc, &codec.CborHandle{})
	return dec.Decode(r)
}
