package object

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
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
	Certificate *Certificate     `nimona:"certificate:o"`
}

func (s *Signature) IsEmpty() bool {
	return s == nil || len(s.X) == 0
}

func (s *Signature) MarshalMap() (Map, error) {
	r := Map{}
	if !s.Signer.IsEmpty() {
		r["signer"] = String(s.Signer.String())
	}
	if s.Alg != "" {
		r["alg"] = String(s.Alg)
	}
	if len(s.X) > 0 {
		r["x"] = Data(s.X)
	}
	if s.Certificate != nil {
		c, err := s.Certificate.MarshalObject()
		if err != nil {
			return nil, err
		}
		r["certificate"] = c
	}
	return r, nil
}

func (s *Signature) UnmarshalMap(m Map) error {
	if t, ok := m["signer"]; ok {
		if v, ok := t.(String); ok {
			k := crypto.PublicKey{}
			if err := k.UnmarshalString(string(v)); err == nil {
				s.Signer = k
			}
		}
	}
	if t, ok := m["alg"]; ok {
		if v, ok := t.(String); ok {
			s.Alg = string(v)
		}
	}
	if t, ok := m["x"]; ok {
		if v, ok := t.(Data); ok {
			s.X = []byte(v)
		}
	}
	if t, ok := m["certificate"]; ok {
		if v, ok := t.(*Object); ok {
			c := &Certificate{}
			err := Unmarshal(v, c)
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
	h, err := NewCID(o)
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
