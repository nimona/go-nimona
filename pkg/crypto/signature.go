package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"

	"nimona.io/pkg/object"
)

var (
	// ErrInvalidBlockType is returned when the signature being verified
	// is not an encoded block of type "signature".
	ErrInvalidBlockType = errors.New("invalid block type")
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

//go:generate go run nimona.io/tools/objectify -schema /signature -type Signature -in signature.go -out signature_generated.go

// Signature block (container), currently supports only ES256
type Signature struct {
	Alg string `json:"alg"`
	R   []byte `json:"r"`
	S   []byte `json:"s"`
}

// NewSignature returns a signature given some bytes and a private key
func NewSignature(key *Key, alg string, o *object.Object) (*Signature, error) {
	if key == nil {
		return nil, errors.New("missing key")
	}

	mKey := key.Materialize()
	if mKey == nil {
		return nil, errors.New("could not materialize")
	}

	pKey, ok := mKey.(*ecdsa.PrivateKey)
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
		hash, err = object.ObjectHash(o)
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
		Alg: alg,
		R:   r.Bytes(),
		S:   s.Bytes(),
	}, nil
}
