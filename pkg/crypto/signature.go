package crypto

import (
	"nimona.io/pkg/errors"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
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
	k PrivateKey,
	o object.Object,
) (*Signature, error) {
	h := hash.New(o)
	x := k.Sign(h.D)
	s := &Signature{
		Signer: NewCertificate(k),
		Alg:    AlgorithmObjectHash,
		X:      x,
	}
	return s, nil
}

func GetObjectSignature(o object.Object) (*Signature, error) {
	so := o.GetSignature()
	if so == nil {
		return nil, errors.New("object is not signed")
	}
	vo := object.Object{}
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

func GetSignatureKeys(sig *Signature) (pks []PublicKey) {
	for {
		if sig == nil || sig.Signer == nil || sig.Signer.Subject == "" {
			return
		}
		pks = append(pks, sig.Signer.Subject)
		sig = sig.Signer.Signature
	}
}
