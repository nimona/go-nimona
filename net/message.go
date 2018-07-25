package net

import (
	"crypto/sha256"
	"errors"
	"reflect"

	"github.com/ugorji/go/codec"
)

// NewEnvelope is a helper function for creating Envelopes
func NewEnvelope(contentType string, recipients []string, payload interface{}) *Envelope {
	// TODO do we need to add the owner on the policy as well?
	envelope := &Envelope{
		Version: 0,
		Headers: Headers{
			Type: contentType,
		},
		Payload: payload,
	}
	subjects := []string{}
	for _, recipient := range recipients {
		// TODO verify valid subject
		if recipient != "" {
			subjects = append(subjects, recipient)
		}
	}
	if len(subjects) > 0 {
		envelope.Headers.Policies = []Policy{
			Policy{
				Description: "policy for recipients",
				Subjects:    subjects,
				Actions:     []string{"read"},
				Effect:      "allow",
			},
		}
	}
	return envelope
}

// Policy for Envelope
type Policy struct {
	Description string   `json:"description,omitempty"`
	Subjects    []string `json:"subjects"`
	Actions     []string `json:"actions"`
	Effect      string   `json:"effect"`
}

// Headers for Envelope
type Headers struct {
	ID       string   `json:"@id,omitempty"`
	Type     string   `json:"@type"`
	Parent   string   `json:"parent_id,omitempty"`
	Policies []Policy `json:"policies,omitempty"`
	Signer   string   `json:"signer,omitempty"`
}

// Envelope for exchanging data via the messenger
type Envelope struct {
	Version   uint        `json:"version"`
	Headers   Headers     `json:"headers,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Signature []byte      `json:"signature,omitempty"`
}

// CodecDecodeSelf helper for cbor unmarshaling
func (envelope *Envelope) CodecDecodeSelf(dec *codec.Decoder) {
	dec.MustDecode(&envelope.Version)
	dec.MustDecode(&envelope.Headers)
	envelope.Payload = GetContentType(envelope.Headers.Type)
	dec.MustDecode(&envelope.Payload)
	dec.MustDecode(&envelope.Signature)
}

// CodecEncodeSelf helper for cbor marshaling
func (envelope *Envelope) CodecEncodeSelf(enc *codec.Encoder) {
	enc.MustEncode(&envelope.Version)
	enc.MustEncode(&envelope.Headers)
	enc.MustEncode(&envelope.Payload)
	enc.MustEncode(&envelope.Signature)
}

// IsSigned checks if the envelope is signed
func (envelope *Envelope) IsSigned() bool {
	// TODO make this part of the envelope and digest?
	return envelope.Signature != nil && len(envelope.Signature) > 0
}

// GetRecipientsFromEnvelopePolicies returns the subjects from all the policies
// of the envelope
func GetRecipientsFromEnvelopePolicies(envelope *Envelope) []string {
	recipients := []string{}
	for _, policy := range envelope.Headers.Policies {
		recipients = append(recipients, policy.Subjects...)
	}
	return recipients
}

func getSignatureDigest(envelope *Envelope) ([]byte, error) {
	headers := envelope.Headers
	headers.ID = ""
	digest := []interface{}{
		// envelope.Version,
		headers,
		envelope.Payload,
	}

	digestBytes, err := Marshal(digest)
	if err != nil {
		return nil, err
	}

	return digestBytes, nil
}

// SetID sets the id of the envelope
func SetID(envelope *Envelope) error {
	id, err := ID(envelope)
	if err != nil {
		return err
	}

	envelope.Headers.ID = id
	return nil
}

// ID calculated the id for the envelope
func ID(envelope *Envelope) (string, error) {
	digest, err := getIdetifierDigest(envelope)
	if err != nil {
		return "", err
	}

	idBytes := sha256.Sum256(digest)
	id := Base58Encode(idBytes[:])
	return "30x" + id, nil
}

func getIdetifierDigest(envelope *Envelope) ([]byte, error) {
	headers := envelope.Headers
	headers.ID = ""
	digest := []interface{}{
		// envelope.Version,
		headers,
		envelope.Payload,
	}

	digestBytes, err := Marshal(digest)
	if err != nil {
		return nil, err
	}

	return digestBytes, nil
}

// SetSigner sets the signer's id on the envelope
func SetSigner(envelope *Envelope, signerPeerInfo *PrivatePeerInfo) {
	envelope.Headers.Signer = signerPeerInfo.ID
}

// Sign envelope given a private peer info
// TODO signer should already be set in the envelope, so maybe we can get
// the keys from the address book?
func Sign(envelope *Envelope, signerPeerInfo *PrivatePeerInfo) error {
	if envelope.Headers.Signer == "" {
		return errors.New("no sigher specified")
	}

	digest, err := getSignatureDigest(envelope)
	if err != nil {
		return err
	}

	signature, err := SignData(digest, signerPeerInfo.PrivateKey)
	if err != nil {
		return err
	}

	envelope.Signature = signature
	return nil
}

// Verify envelope's signature
func (envelope *Envelope) Verify() error {
	digest, err := getSignatureDigest(envelope)
	if err != nil {
		return err
	}

	return Verify(envelope.Headers.Signer, digest, envelope.Signature)
}

// Marshal into cbor
func Marshal(o interface{}) ([]byte, error) {
	b := []byte{}
	enc := codec.NewEncoderBytes(&b, getCborHandler())
	if err := enc.Encode(o); err != nil {
		return nil, err
	}

	return b, nil
}

// Unmarshal from cbor
func Unmarshal(b []byte) (*Envelope, error) {
	m := &Envelope{}
	dec := codec.NewDecoderBytes(b, getCborHandler())

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return m, nil
}

func getCborHandler() codec.Handle {
	ch := &codec.CborHandle{}
	ch.Canonical = true
	ch.MapType = reflect.TypeOf(map[string]interface{}(nil))
	return ch
}
