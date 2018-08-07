package blocks

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
	// ID        string   `json:"@id,omitempty"`
	Type      string   `json:"@type"`
	Parent    string   `json:"parent_id,omitempty"`
	Ephemeral bool     `json:"ephemeral,omitempty"`
	Policies  []Policy `json:"policies,omitempty"`
	Signer    string   `json:"signer,omitempty"`
}

// Block for exchanging data via the exchange
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

// ID calculated the id for the block
func (block *Block) ID() (string, error) {
	return ID(block)
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

func GetSignatureDigest(block *Block) ([]byte, error) {
	meta := block.Metadata
	digest := []interface{}{
		meta,
		block.Payload,
	}

	digestBytes, err := Marshal(digest)
	if err != nil {
		return nil, err
	}

	return digestBytes, nil
}

// ID calculated the id for the block
func ID(block *Block) (string, error) {
	digest, err := GetSignatureDigest(block)
	if err != nil {
		return "", err
	}

	idBytes := sha256.Sum256(digest)
	id := Base58Encode(idBytes[:])
	return "30x" + id, nil
}

// Marshal into cbor
func Marshal(o interface{}) ([]byte, error) {
	b := []byte{}
	enc := codec.NewEncoderBytes(&b, CborHandler())
	if err := enc.Encode(o); err != nil {
		return nil, err
	}

	return b, nil
}

// Unmarshal from cbor
func Unmarshal(b []byte) (*Block, error) {
	m := &Block{}
	dec := codec.NewDecoderBytes(b, CborHandler())

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return m, nil
}

// CborHandler for un/marshaling blocks
func CborHandler() codec.Handle {
	ch := &codec.CborHandle{}
	ch.Canonical = true
	ch.MapType = reflect.TypeOf(map[string]interface{}(nil))
	return ch
}

// BestEffortID returns an error-free ID
// TODO can we instead of this, make ID "error free"?
func BestEffortID(block *Block) string {
	blockID, _ := block.ID()
	if blockID == "" {
		return "<invalid-block-id>"
	}

	return blockID
}

// Copy a block
func Copy(block *Block) *Block {
	b, _ := Marshal(block)
	newBlock, _ := Unmarshal(b)
	return newBlock
}
