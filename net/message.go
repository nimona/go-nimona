package net

import (
	"encoding/json"
	"reflect"

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

func (message *Message) CodecDecodeSelf(dec *codec.Decoder) {
	dec.MustDecode(&message.Version)
	dec.MustDecode(&message.Type)
	message.Payload = GetContentType(message.Type)
	dec.MustDecode(&message.Payload)
	dec.MustDecode(&message.Headers)
	dec.MustDecode(&message.Payload)
	dec.MustDecode(&message.Signature)
}

func (message *Message) CodecEncodeSelf(enc *codec.Encoder) {
	enc.MustEncode(&message.Version)
	enc.MustEncode(&message.Type)
	enc.MustEncode(&message.Payload)
	enc.MustEncode(&message.Headers)
	enc.MustEncode(&message.Payload)
	enc.MustEncode(&message.Signature)
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
	enc := codec.NewEncoderBytes(&b, getCborHandler())
	if err := enc.Encode(o); err != nil {
		return nil, err
	}

	return b, nil
}

func Unmarshal(b []byte) (*Message, error) {
	m := &Message{}
	dec := codec.NewDecoderBytes(b, getCborHandler())

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return m, nil
}

// DecodePayload decodes the message's payload according to the coded,
// and stores the result in the value pointed to by r.
func (h *Message) DecodePayload(r interface{}) error {
	enc, err := Marshal(h.Payload)
	if err != nil {
		return err
	}

	dec := codec.NewDecoderBytes(enc, getCborHandler())
	return dec.Decode(r)
}

func PrettifyMessage(message *Message) string {
	b, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		return "[cannot marshal, " + err.Error() + "]"
	}

	return string(b)
}

func getCborHandler() codec.Handle {
	ch := &codec.CborHandle{}
	ch.Canonical = true
	ch.MapType = reflect.TypeOf(map[string]interface{}(nil))
	return ch
}
