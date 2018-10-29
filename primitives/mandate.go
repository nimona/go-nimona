package primitives

import (
	"errors"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	ucodec "github.com/ugorji/go/codec"
)

type Mandate struct {
	Subject   *Key       `json:"subject" mapstructure:"-"`
	Policy    Policy     `json:"policy" mapstructure:"policy,omitempty"`
	Signature *Signature `json:"-" mapstructure:"-"`
}

func (s *Mandate) Block() *Block {
	p := structs.New(s)
	p.TagName = "mapstructure"
	m := p.Map()
	m["subject"] = s.Subject
	return &Block{
		Type:      "nimona.io/mandate",
		Payload:   m,
		Signature: s.Signature,
	}
}

func (s *Mandate) FromBlock(block *Block) {
	p := struct {
		Subject *Block `json:"subject" mapstructure:"subject,omitempty"`
		Policy  Policy `json:"policy" mapstructure:"policy,omitempty"`
	}{}
	if err := mapstructure.Decode(block.Payload, &p); err != nil {
		panic(err)
	}
	s.Policy = p.Policy
	s.Subject = &Key{}
	s.Subject.FromBlock(p.Subject)
	s.Signature = block.Signature
}

// CodecDecodeSelf helper for cbor unmarshaling
func (s *Mandate) CodecDecodeSelf(dec *ucodec.Decoder) {
	b := &Block{}
	dec.MustDecode(b)
	s.FromBlock(b)
}

// CodecEncodeSelf helper for cbor marshaling
func (s *Mandate) CodecEncodeSelf(enc *ucodec.Encoder) {
	b := s.Block()
	enc.MustEncode(b)
}

// NewMandate returns a signed mandate given an authority key, a subject key,
// and a policy
func NewMandate(authority, subject *Key, policy Policy) (*Mandate, error) {
	if authority == nil {
		return nil, errors.New("missing authority")
	}

	if subject == nil {
		return nil, errors.New("missing subject")
	}

	m := &Mandate{
		Subject: subject,
		Policy:  policy,
	}

	b := m.Block()
	if err := Sign(b, authority); err != nil {
		return nil, err
	}

	m.Signature = b.Signature
	return m, nil
}
