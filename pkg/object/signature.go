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

// NewSignature returns a signature given some bytes and a private key
func NewSignature(
	k crypto.PrivateKey,
	o Object,
) (*Signature, error) {
	h := NewHash(o)
	x := k.Sign(h.Bytes())
	s := &Signature{
		Signer: k.PublicKey(),
		Alg:    AlgorithmObjectHash,
		X:      x,
	}
	return s, nil
}

func GetObjectSignature(o Object) (*Signature, error) {
	so := o.GetSignature()
	if so == nil {
		return nil, errors.New("object is not signed")
	}
	vo := Object{}
	if err := vo.FromMap(*so); err != nil {
		return nil, errors.Wrap(
			errors.New("invalid signature object"),
			err,
		)
	}
	s := &Signature{}
	if err := s.FromObject(vo); err != nil {
		return nil, errors.Wrap(
			errors.New("invalid signature"),
			err,
		)
	}
	return s, nil
}
