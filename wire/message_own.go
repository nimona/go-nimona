package wire

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"fmt"

	"github.com/nimona/go-nimona/peer"
	"github.com/ugorji/go/codec"
)

func NewMessage(contentType string, recipients []string, payload interface{}) (*Message, error) {
	message := &Message{
		Headers: Headers{
			ContentType: contentType,
			Recipients:  recipients,
		},
	}
	if err := message.EncodePayload(payload); err != nil {
		return nil, err
	}
	return message, nil
}

type Headers struct {
	ContentType string
	Recipients  []string
	Signer      []byte
}

// Message for exchanging data via the wire
type Message struct {
	Version   int
	Headers   Headers
	Payload   []byte
	Signature []byte
}

func (message *Message) IsSigned() bool {
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

	digestHash := sha512.Sum512(digestBytes)
	// fmt.Printf("DIGEST: %x\n", digestHash[:])
	// asdfsdf, _ := json.MarshalIndent(digest, "", "  ")
	// fmt.Println(string(asdfsdf))
	return digestHash[:], nil
}

func (message *Message) Sign(signerPeerInfo *peer.SecretPeerInfo) error {
	message.Headers.Signer = signerPeerInfo.PublicKey
	digest, err := getMessageDigest(message)
	if err != nil {
		return err
	}

	key := signerPeerInfo.GetSecretKey()
	signatureBody, err := rsa.SignPSS(rand.Reader, key, crypto.SHA512, digest, &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA512,
	})
	if err != nil {
		return fmt.Errorf("rsa.SignPSS error %s", err)
	}

	message.Signature = signatureBody
	return nil
}

func (message *Message) Verify() error {
	publicKey, err := x509.ParsePKCS1PublicKey(message.Headers.Signer)
	if err != nil {
		return err
	}

	digest, err := getMessageDigest(message)
	if err != nil {
		return err
	}

	return rsa.VerifyPSS(publicKey, crypto.SHA512, digest, message.Signature, &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA512,
	})
}

func Marshal(o interface{}) ([]byte, error) {
	b := []byte{}
	enc := codec.NewEncoderBytes(&b, &codec.CborHandle{})
	if err := enc.Encode(o); err != nil {
		return nil, err
	}

	return b, nil
}

func Unmarshal(b []byte) (o interface{}, err error) {
	dec := codec.NewDecoderBytes(b, &codec.CborHandle{})
	if err := dec.Decode(&o); err != nil {
		return nil, err
	}

	return o, nil
}

// DecodePayload decodes the message's payload according to the coded,
// and stores the result in the value pointed to by r.
func (h *Message) DecodePayload(r interface{}) error {
	dec := codec.NewDecoderBytes(h.Payload, &codec.CborHandle{})
	return dec.Decode(r)
}

// EncodePayload encodes the given value using the message's codec, and stores
// the result in the message's payload.
func (h *Message) EncodePayload(r interface{}) error {
	payloadBytes := []byte{}
	enc := codec.NewEncoderBytes(&payloadBytes, &codec.CborHandle{})
	if err := enc.Encode(r); err != nil {
		return err
	}
	h.Payload = payloadBytes
	return nil
}
