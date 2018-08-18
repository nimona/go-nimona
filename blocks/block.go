package blocks

import (
	"reflect"

	"github.com/ugorji/go/codec"
)

// NewEphemeralBlock is a helper function for creating ephemeral Blocks.
func NewEphemeralBlock(contentType string, payload interface{}, recipients ...string) *Block {
	block := NewBlock(contentType, payload, recipients...)
	block.Metadata.Ephemeral = true
	return block
}

// NewBlock is a helper function for creating Blocks.
func NewBlock(contentType string, payload interface{}, recipients ...string) *Block {
	// TODO do we need to add the owner on the policy as well?
	block := &Block{
		Type:    contentType,
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
	Parent    string   `json:"parentID,omitempty"`
	Ephemeral bool     `json:"ephemeral,omitempty"`
	Policies  []Policy `json:"policies,omitempty"`
	Signer    string   `json:"signer,omitempty"`
}

// Block for exchanging data via the exchange
type Block struct {
	Type      string            `json:"type,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Metadata  Metadata          `json:"metadata,omitempty"`
	Payload   interface{}       `json:"payload,omitempty"`
	Signature []byte            `json:"signature,omitempty"`
}

type _block struct {
	Type      string            `json:"type,omitempty"`
	Headers   map[string]string `json:"head,omitempty"`
	Metadata  *Metadata         `json:"meta,omitempty"`
	Signature []byte            `json:"sign,omitempty"`
	Payload   interface{}       `json:"data,omitempty"`
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
	b := &_block{}
	dec.MustDecode(b)

	var payload interface{}
	if p := GetContentType(b.Type); p != nil {
		payload = p
	}

	block.Type = b.Type
	block.Headers = b.Headers
	if b.Metadata != nil {
		block.Metadata = *b.Metadata
	}
	block.Signature = b.Signature

	pb, _ := Marshal(b.Payload)
	Unmarshal(pb, &payload)
	block.Payload = payload
}

// CodecEncodeSelf helper for cbor marshaling
func (block *Block) CodecEncodeSelf(enc *codec.Encoder) {
	p := toThinBlock(block, true, true)
	enc.MustEncode(p)
}

func toThinBlock(block *Block, headers, signature bool) *_block {
	p := &_block{
		Type:     block.Type,
		Metadata: nil,
		Payload:  block.Payload,
	}
	if headers {
		p.Headers = block.Headers
	}
	if signature {
		p.Signature = block.Signature
	}
	if block.Metadata.Parent != "" ||
		block.Metadata.Ephemeral != false ||
		len(block.Metadata.Policies) > 0 ||
		block.Metadata.Signer != "" {
		p.Metadata = &block.Metadata
	}

	return p
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

// GetSignatureDigest returns a marshaled version of the block, without
// headers and signature. Used for consistent hash/ID.
func GetSignatureDigest(block *Block) ([]byte, error) {
	cleanBlock := toThinBlock(block, false, false)
	digestBytes, err := Marshal(cleanBlock)
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

	return SumSha3(digest)
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
func Unmarshal(b []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(b, CborHandler())
	if err := dec.Decode(v); err != nil {
		return err
	}

	return nil
}

// CborHandler for un/marshaling blocks
func CborHandler() *codec.CborHandle {
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
	v := &Block{}
	b, _ := Marshal(block)
	Unmarshal(b, v)
	return v
}
