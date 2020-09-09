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
	Signer      crypto.PublicKey `json:"signer:s,omitempty" mapstructure:"signer:s,omitempty"`
	Alg         string           `json:"alg:s,omitempty" mapstructure:"alg:s,omitempty"`
	X           []byte           `json:"x:d,omitempty" mapstructure:"x:d,omitempty"`
	Certificate *Certificate     `json:"certificate:m,omitempty" mapstructure:"certificate:m,omitempty"`
}

func (s Signature) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"signer:s": s.Signer.String(),
		"alg:s":    s.Alg,
		"x:d":      s.X,
	}
	if s.Certificate != nil {
		m["certificate:m"] = s.Certificate.ToObject().ToMap()
	}
	return m
}

// NewSignature returns a signature given some bytes and a private key
func NewSignature(
	k crypto.PrivateKey,
	o Object,
) (Signature, error) {
	x := k.Sign(o.Hash().rawBytes())
	s := Signature{
		Signer: k.PublicKey(),
		Alg:    AlgorithmObjectHash,
		X:      x,
	}
	return s, nil
}
