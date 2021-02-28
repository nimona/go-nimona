package object

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
)

var (
	// ErrInvalidObjectType is returned when the signature being verified
	// is not an encoded object of type "signature".
	ErrInvalidObjectType = errors.New("invalid object type")
	// ErrAlgorithNotImplemented is returned when the algorithm specified
	// has not been implemented
	ErrAlgorithNotImplemented = errors.New("algorithm not implemented")
)

const (
	// AlgorithmObjectCID for creating ObjectCID+ES256 based signatures
	AlgorithmObjectCID = "OH_ES256"
)

type Signature struct {
	Signer      crypto.PublicKey `nimona:"signer:s,omitempty"`
	Alg         string           `nimona:"alg:s,omitempty"`
	X           []byte           `nimona:"x:d,omitempty"`
	Certificate *Certificate     `nimona:"certificate:m,omitempty"`
}

func (s Signature) IsEmpty() bool {
	return len(s.X) == 0
}

func (s Signature) Map() Map {
	r := Map{}
	if !s.Signer.IsEmpty() {
		r["signer"] = String(s.Signer)
	}
	if s.Alg != "" {
		r["alg"] = String(s.Alg)
	}
	if len(s.X) > 0 {
		r["x"] = Data(s.X)
	}
	if s.Certificate != nil {
		r["certificate"] = s.Certificate.ToObject()
	}
	return r
}

func SignatureFromMap(m Map) Signature {
	r := Signature{}
	if t, ok := m["signer"]; ok {
		if s, ok := t.(String); ok {
			r.Signer = crypto.PublicKey(s)
		}
	}
	if t, ok := m["alg"]; ok {
		if s, ok := t.(String); ok {
			r.Alg = string(s)
		}
	}
	if t, ok := m["alg"]; ok {
		if s, ok := t.(String); ok {
			r.Alg = string(s)
		}
	}
	if t, ok := m["x"]; ok {
		if s, ok := t.(Data); ok {
			r.X = []byte(s)
		}
	}
	if t, ok := m["certificate"]; ok {
		if s, ok := t.(*Object); ok {
			c := &Certificate{}
			if err := c.FromObject(s); err == nil {
				r.Certificate = c
			}
		}
	}
	return r
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
