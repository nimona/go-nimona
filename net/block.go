package net

import (
	"crypto/sha256"
	"reflect"

	"github.com/ugorji/go/codec"
)

func NewEphemeralBlock(contentType string, payload interface{}, recipients ...string) *Block {
	block := NewBlock(contentType, payload, recipients...)
	block.Metadata.Ephemeral = true
	return block
}

// NewBlock is a helper function for creating Blocks
func NewBlock(contentType string, payload interface{}, recipients ...string) *Block {
	// TODO do we need to add the owner on the policy as well?
	block := &Block{
		Version: 0,
		Metadata: Metadata{
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
		block.Metadata.Policies = []Policy{
			Policy{
				Description: "policy for recipients",
				Subjects:    subjects,
				Actions:     []string{"read"},
				Effect:      "allow",
			},
		}
	}
	return block
}

// Policy for Block
type Policy struct {
	Description string   `json:"description,omitempty"`
	Subjects    []string `json:"subjects"`
	Actions     []string `json:"actions"`
	Effect      string   `json:"effect"`
}

// Metadata for Block
type Metadata struct {
	ID        string   `json:"@id,omitempty"`
	Type      string   `json:"@type"`
	Parent    string   `json:"parent_id,omitempty"`
	Ephemeral bool     `json:"ephemeral,omitempty"`
	Policies  []Policy `json:"policies,omitempty"`
	Signer    string   `json:"signer,omitempty"`
}

// Block for exchanging data via the messenger
type Block struct {
	Version   uint              `json:"version"`
	Headers   map[string]string `json:"headers,omitempty"`
	Metadata  Metadata          `json:"metadata,omitempty"`
	Payload   interface{}       `json:"payload,omitempty"`
	Signature []byte            `json:"signature,omitempty"`
}

// SetHeader pair in block
func (block *Block) SetHeader(k, v string) {
	if block.Headers == nil {
		block.Headers = map[string]string{}
	}
	block.Headers[k] = v
}

// GetHeader by key
func (block *Block) GetHeader(k string) string {
	if block.Headers == nil {
		return ""
	}
	return block.Headers[k]
}

// CodecDecodeSelf helper for cbor unmarshaling
func (block *Block) CodecDecodeSelf(dec *codec.Decoder) {
	dec.MustDecode(&block.Version)
	dec.MustDecode(&block.Metadata)
	block.Payload = GetContentType(block.Metadata.Type)
	dec.MustDecode(&block.Payload)
	dec.MustDecode(&block.Signature)
	dec.MustDecode(&block.Headers)
}

// CodecEncodeSelf helper for cbor marshaling
func (block *Block) CodecEncodeSelf(enc *codec.Encoder) {
	enc.MustEncode(&block.Version)
	enc.MustEncode(&block.Metadata)
	enc.MustEncode(&block.Payload)
	enc.MustEncode(&block.Signature)
	enc.MustEncode(&block.Headers)
}

// IsSigned checks if the block is signed
func (block *Block) IsSigned() bool {
	// TODO make this part of the block and digest?
	return block.Signature != nil && len(block.Signature) > 0
}

// GetRecipientsFromBlockPolicies returns the subjects from all the policies
// of the block
func GetRecipientsFromBlockPolicies(block *Block) []string {
	recipients := []string{}
	for _, policy := range block.Metadata.Policies {
		recipients = append(recipients, policy.Subjects...)
	}
	return recipients
}

func getSignatureDigest(block *Block) ([]byte, error) {
	headers := block.Metadata
	headers.ID = ""
	digest := []interface{}{
		// block.Version,
		headers,
		block.Payload,
	}

	digestBytes, err := Marshal(digest)
	if err != nil {
		return nil, err
	}

	return digestBytes, nil
}

// SetID sets the id of the block
func SetID(block *Block) error {
	id, err := ID(block)
	if err != nil {
		return err
	}

	block.Metadata.ID = id
	return nil
}

// ID calculated the id for the block
func ID(block *Block) (string, error) {
	digest, err := getIdetifierDigest(block)
	if err != nil {
		return "", err
	}

	idBytes := sha256.Sum256(digest)
	id := Base58Encode(idBytes[:])
	return "30x" + id, nil
}

func getIdetifierDigest(block *Block) ([]byte, error) {
	headers := block.Metadata
	headers.ID = ""
	digest := []interface{}{
		// block.Version,
		headers,
		block.Payload,
	}

	digestBytes, err := Marshal(digest)
	if err != nil {
		return nil, err
	}

	return digestBytes, nil
}

// SetSigner sets the signer's id on the block
func SetSigner(block *Block, signerPeerInfo *PrivatePeerInfo) {
	block.Metadata.Signer = signerPeerInfo.ID
}

// Sign block given a private peer info
// TODO signer should already be set in the block, so maybe we can get
// the keys from the address book?
func Sign(block *Block, signerPeerInfo *PrivatePeerInfo) error {
	// if block.Metadata.Signer == "" {
	// 	return errors.New("no sigher specified")
	// }
	block.Metadata.Signer = signerPeerInfo.ID

	digest, err := getSignatureDigest(block)
	if err != nil {
		return err
	}

	signature, err := SignData(digest, signerPeerInfo.PrivateKey)
	if err != nil {
		return err
	}

	block.Signature = signature

	return SetID(block)
}

// Verify block's signature
func (block *Block) Verify() error {
	if len(block.Signature) == 0 {
		return nil
	}

	digest, err := getSignatureDigest(block)
	if err != nil {
		return err
	}

	return Verify(block.Metadata.Signer, digest, block.Signature)
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
func Unmarshal(b []byte) (*Block, error) {
	m := &Block{}
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
