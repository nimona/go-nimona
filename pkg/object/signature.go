package object

import (
	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object/cid"
)

const (
	// ErrInvalidObjectType is returned when the signature being verified
	// is not an encoded object of type "signature".
	ErrInvalidObjectType = errors.Error("invalid object type")
	// ErrAlgorithNotImplemented is returned when the algorithm specified
	// has not been implemented
	ErrAlgorithNotImplemented = errors.Error("algorithm not implemented")
)

const (
	// AlgorithmObjectCID for creating ObjectCID+ES256 based signatures
	AlgorithmObjectCID = "OH_ES256"
)

type Signature struct {
	Signer      crypto.PublicKey `nimona:"signer:s"`
	Alg         string           `nimona:"alg:s"`
	X           []byte           `nimona:"x:d"`
	Certificate *Certificate     `nimona:"certificate:m"`
}

func (s *Signature) IsEmpty() bool {
	return s == nil || len(s.X) == 0
}

func (s *Signature) MarshalMap() (chore.Map, error) {
	r := chore.Map{}
	if !s.Signer.IsEmpty() {
		r["signer"] = chore.String(s.Signer.String())
	}
	if s.Alg != "" {
		r["alg"] = chore.String(s.Alg)
	}
	if len(s.X) > 0 {
		r["x"] = chore.Data(s.X)
	}
	if s.Certificate != nil {
		c, err := s.Certificate.MarshalObject()
		if err != nil {
			return nil, err
		}
		m, err := c.MarshalMap()
		if err != nil {
			return nil, err
		}
		r["certificate"] = m
	}
	return r, nil
}

func (s *Signature) UnmarshalMap(m chore.Map) error {
	if t, ok := m["signer"]; ok {
		if v, ok := t.(chore.String); ok {
			k := crypto.PublicKey{}
			if err := k.UnmarshalString(string(v)); err == nil {
				s.Signer = k
			}
		}
	}
	if t, ok := m["alg"]; ok {
		if v, ok := t.(chore.String); ok {
			s.Alg = string(v)
		}
	}
	if t, ok := m["x"]; ok {
		if v, ok := t.(chore.Data); ok {
			s.X = []byte(v)
		}
	}
	if t, ok := m["certificate"]; ok {
		if m, ok := t.(chore.Map); ok {
			o := &Object{}
			err := o.UnmarshalMap(m)
			if err != nil {
				return err
			}
			c := &Certificate{}
			err = Unmarshal(o, c)
			if err != nil {
				return err
			}
			s.Certificate = c
		}
	}
	return nil
}

// NewSignature returns a signature given some bytes and a private key
func NewSignature(
	k crypto.PrivateKey,
	o *Object,
) (Signature, error) {
	m, err := o.MarshalMap()
	if err != nil {
		return Signature{}, err
	}
	h, err := cid.New(m)
	if err != nil {
		return Signature{}, err
	}
	x := k.Sign([]byte(h))
	s := Signature{
		Signer: k.PublicKey(),
		Alg:    AlgorithmObjectCID,
		X:      x,
	}
	return s, nil
}
