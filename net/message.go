package net

import (
	"fmt"

	"github.com/ugorji/go/codec"
)

func NewMessage(contentType string, recipients []string, payload interface{}) (*Message, error) {
	message := &Message{
		Headers: Headers{
			ContentType: contentType,
			Recipients:  recipients,
		},
		Payload: payload,
	}
	return message, nil
}

type Headers struct {
	ContentType string
	Recipients  []string
	Signer      string
}

// Message for exchanging data via the messenger
type Message struct {
	Version   int
	Headers   Headers
	Payload   interface{}
	Signature []byte
}

func (message *Message) IsSigned() bool {
	// TODO make this part of the message and digest?
	return message.Signature != nil && len(message.Signature) > 0
}

func getMessageDigest(message *Message) ([]byte, error) {
	digest := []interface{}{
		message.Version,
		message.Headers,
		message.Payload,
	}

	digestBytes, err := Marshal(digest)
	if err != nil {
		return nil, err
	}

	// digestHash := sha512.Sum512(digestBytes)
	// fmt.Printf("DIGEST: %x\n", digestHash[:])
	// asdfsdf, _ := json.MarshalIndent(digest, "", "  ")
	// fmt.Println(string(asdfsdf))
	// return digestHash[:], nil
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
		fmt.Println("5a")
		return nil, err
	}
	ct := GetContentType(m.Headers.ContentType)
	if err := m.DecodePayload(&ct); err != nil {
		fmt.Println("5b")
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
		fmt.Println("5c")
		return err
	}

	dec := codec.NewDecoderBytes(enc, &codec.CborHandle{})
	return dec.Decode(r)

	// dec := codec.NewDecoderBytes(h.Payload, &codec.CborHandle{})
	// return dec.Decode(r)
	// return nil
}

// // EncodePayload encodes the given value using the message's codec, and stores
// // the result in the message's payload.
// func (h *Message) EncodePayload(r interface{}) error {
// 	// payloadBytes := []byte{}
// 	// enc := codec.NewEncoderBytes(&payloadBytes, &codec.CborHandle{})
// 	// if err := enc.Encode(r); err != nil {
// 	// 	return err
// 	// }
// 	// h.Payload = payloadBytes
// 	return nil
// }
