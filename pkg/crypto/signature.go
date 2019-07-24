package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"

	"nimona.io/internal/errors"
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
	// AlgorithmES256 for creating ES256 based signatures
	AlgorithmES256 = "ES256"
	// AlgorithmObjectHash for creating ObjectHash+ES256 based signatures
	AlgorithmObjectHash = "OH_ES256"
)

//go:generate $GOBIN/objectify -schema /signature -type Signature -in signature.go -out signature_generated.go

// Signature object (container), currently supports only ES256
type Signature struct {
	PublicKey *PublicKey `json:"pub:o"`
	Alg       string     `json:"alg:s"`
	R         []byte     `json:"r:d"`
	S         []byte     `json:"s:d"`
}

// NewSignature returns a signature given some bytes and a private key
func NewSignature(
	key *PrivateKey,
	alg string,
	o object.Object,
) (*Signature, error) {

	if key == nil {
		return nil, errors.New("missing key")
	}

	pKey, ok := key.Key().(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("only ecdsa private keys are currently supported")
	}

	var (
		hash []byte
		err  error
	)

	switch alg {
	// case AlgorithmES256:
	// 	o, err := object.NewFromStruct(v)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	b, err := object.Marshal(o)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	m := map[string]interface{}{}
	// 	if err := object.UnmarshalSimple(b, &m); err != nil {
	// 		return nil, err
	// 	}

	// 	// TODO replace ES256 with OH that should deal with removing the @sig
	// 	delete(m, "@signature")

	// 	b, err = object.Marshal(m)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	h := sha256.Sum256(b)
	// 	hash = h[:]
	case AlgorithmObjectHash:
		hash, err = object.ObjectHashWithoutSignature(o)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrAlgorithNotImplemented
	}

	r, s, err := ecdsa.Sign(rand.Reader, pKey, hash)
	if err != nil {
		return nil, err
	}

	return &Signature{
		PublicKey: key.PublicKey,
		Alg:       alg,
		R:         r.Bytes(),
		S:         s.Bytes(),
	}, nil
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

func GetObjectKeys(o object.Object) (pks []*PublicKey) {
	sig, _ := GetObjectSignature(o)
	for {
		if sig == nil || sig.PublicKey == nil {
			return
		}
		pk := sig.PublicKey
		pks = append(pks, pk)
		sig = pk.Signature
	}
}

func GetSignatureKeys(sig *Signature) (pks []*PublicKey) {
	for {
		if sig == nil || sig.PublicKey == nil {
			return
		}
		pk := sig.PublicKey
		pks = append(pks, pk)
		sig = pk.Signature
	}
}
