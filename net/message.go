package net

import (
	"reflect"

	"github.com/ugorji/go/codec"
)

// NewEnvelope is a helper function for creating Envelopes
func NewEnvelope(contentType string, recipients []string, payload interface{}) (*Envelope, error) {
	envelope := &Envelope{
		Version: 0,
		Type:    contentType,
		Headers: Headers{
			Recipients: recipients,
		},
		Payload: payload,
	}
	return envelope, nil
}

// Headers for Envelope
type Headers struct {
	Recipients []string `json:"recipients,omitempty"`
	Signer     string   `json:"signer,omitempty"`
}

// Envelope for exchanging data via the messenger
type Envelope struct {
	Version   uint        `json:"version"`
	Type      string      `json:"@type"`
	Headers   Headers     `json:"headers,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Signature []byte      `json:"signature,omitempty"`
}

// CodecDecodeSelf helper for cbor unmarshaling
func (envelope *Envelope) CodecDecodeSelf(dec *codec.Decoder) {
	dec.MustDecode(&envelope.Version)
	dec.MustDecode(&envelope.Type)
	envelope.Payload = GetContentType(envelope.Type)
	dec.MustDecode(&envelope.Payload)
	dec.MustDecode(&envelope.Headers)
	dec.MustDecode(&envelope.Payload)
	dec.MustDecode(&envelope.Signature)
}

// CodecEncodeSelf helper for cbor marshaling
func (envelope *Envelope) CodecEncodeSelf(enc *codec.Encoder) {
	enc.MustEncode(&envelope.Version)
	enc.MustEncode(&envelope.Type)
	enc.MustEncode(&envelope.Payload)
	enc.MustEncode(&envelope.Headers)
	enc.MustEncode(&envelope.Payload)
	enc.MustEncode(&envelope.Signature)
}

// IsSigned checks if the envelope is signed
func (envelope *Envelope) IsSigned() bool {
	// TODO make this part of the envelope and digest?
	return envelope.Signature != nil && len(envelope.Signature) > 0
}

func getEnvelopeDigest(envelope *Envelope) ([]byte, error) {
	digest := []interface{}{
		envelope.Version,
		envelope.Type,
		envelope.Headers,
		envelope.Payload,
	}

	digestBytes, err := Marshal(digest)
	if err != nil {
		return nil, err
	}

	return digestBytes, nil
}

// Sign envelope given a private peer info
func (envelope *Envelope) Sign(signerPeerInfo *PrivatePeerInfo) error {
	envelope.Headers.Signer = signerPeerInfo.ID
	digest, err := getEnvelopeDigest(envelope)
	if err != nil {
		return err
	}

	signature, err := Sign(digest, signerPeerInfo.PrivateKey)
	if err != nil {
		return err
	}

	envelope.Signature = signature
	return nil
}

// Verify envelope's signature
func (envelope *Envelope) Verify() error {
	digest, err := getEnvelopeDigest(envelope)
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

// DecodePayload decodes the envelope's payload according to the coded,
// and stores the result in the value pointed to by r.
func (envelope *Envelope) DecodePayload(r interface{}) error {
	enc, err := Marshal(envelope.Payload)
	if err != nil {
		return err
	}

	dec := codec.NewDecoderBytes(enc, getCborHandler())
	return dec.Decode(r)
}

func getCborHandler() codec.Handle {
	ch := &codec.CborHandle{}
	ch.Canonical = true
	ch.MapType = reflect.TypeOf(map[string]interface{}(nil))
	return ch
}
