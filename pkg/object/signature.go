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
	// AlgorithmObjectHash for creating ObjectHash+ES256 based signatures
	AlgorithmObjectHash = "OH_ES256"
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

func SignatureToMap(s *Signature) map[string]interface{} {
	if s == nil || s.IsEmpty() {
		return nil
	}
	r := map[string]interface{}{}
	if !s.Signer.IsEmpty() {
		r["signer:s"] = s.Signer.String()
	}
	if s.Alg != "" {
		r["alg:s"] = s.Alg
	}
	if len(s.X) > 0 {
		r["x:d"] = s.X
	}
	if s.Certificate != nil {
		r["certificate:m"] = s.Certificate.ToObject().ToMap()
	}
	return r
}

// NewSignature returns a signature given some bytes and a private key
func NewSignature(
	k crypto.PrivateKey,
	o *Object,
) (Signature, error) {
	h, err := NewHash(o)
	if err != nil {
		return Signature{}, err
	}
	x := k.Sign([]byte(h))
	s := Signature{
		Signer: k.PublicKey(),
		Alg:    AlgorithmObjectHash,
		X:      x,
	}
	return s, nil
}
