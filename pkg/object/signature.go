package object

import (
	"reflect"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/did"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/tilde"
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
	// AlgorithmObjectHash is the only supported signing algorithm right now.
	AlgorithmObjectHash = "EdDSA"
)

type Signature struct {
	_         *Metadata        `nimona:"@metadata:m,type=Signature"`
	Delegator did.DID          `nimona:"d:s"`
	Signer    did.DID          `nimona:"s:s"`
	Key       crypto.PublicKey `nimona:"jwk:s"`
	Alg       string           `nimona:"alg:s"`
	X         []byte           `nimona:"x:d"`
}

func (s *Signature) IsEmpty() bool {
	return s == nil || len(s.X) == 0
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
	s, err := newSignature(k, m)
	if err != nil {
		return Signature{}, err
	}
	return *s, nil
}

func newSignature(
	k crypto.PrivateKey,
	m tilde.Map,
) (*Signature, error) {
	h, err := m.Hash().Bytes()
	if err != nil {
		return nil, err
	}
	x := k.Sign(h)
	s := &Signature{
		Signer: k.PublicKey().DID(),
		Key:    k.PublicKey(),
		Alg:    AlgorithmObjectHash,
		X:      x,
	}
	return s, nil
}

// Sign an object given a private key, updates the object's metadata in place
func Sign(k crypto.PrivateKey, o *Object) error {
	s, err := NewSignature(k, o)
	if err != nil {
		return err
	}
	o.Metadata.Signature = s
	return nil
}

// SignDeep an object and all nested objects we own or have no owner
// WARNING: THIS _WILL_ CHANGE, DO NOT USE!
// TODO: not sure which nested objects this should sign. All? Own?
func SignDeep(k crypto.PrivateKey, o *Object) error {
	var signErr error
	pk := k.PublicKey()
	Traverse(o, func(path string, v interface{}) bool {
		m, ok := v.(tilde.Map)
		if !ok {
			return true
		}
		if !isObject(m) {
			return true
		}
		meta := &Metadata{}
		mmeta, ok := m["@metadata"].(tilde.Map)
		if !ok {
			return true
		}
		err := unmarshalMap(tilde.MapHint, mmeta, reflect.ValueOf(meta))
		if err != nil {
			return true
		}
		if !meta.Signature.IsEmpty() {
			return true
		}
		if meta.Owner == did.Empty {
			return true
		}
		// TODO(geoah): figure out if we should be signing this object
		if !meta.Owner.Equals(pk.DID()) {
			return true
		}
		sig, err := newSignature(k, m)
		if err != nil {
			signErr = err
			return true
		}
		osig, err := Marshal(sig)
		if err != nil {
			return true
		}
		msig, err := osig.MarshalMap()
		if err != nil {
			return true
		}
		mmeta["_signature"] = msig
		return true
	})
	// TODO(geoah): figure out if we should be signing this object
	if !o.Metadata.Owner.Equals(pk.DID()) {
		return nil
	}
	m, err := o.MarshalMap()
	if err != nil {
		return err
	}
	s, err := newSignature(k, m)
	if err != nil {
		return err
	}
	o.Metadata.Signature = *s
	return signErr
}

func isObject(m tilde.Map) bool {
	if _, ok := m["@type"]; ok {
		return true
	}
	return false
}
