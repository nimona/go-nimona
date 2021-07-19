package object

import (
	"reflect"

	"nimona.io/pkg/chore"
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
	// AlgorithmObjectHash for creating ObjectHash+ES256 based signatures
	AlgorithmObjectHash = "OH_ES256"
)

type Signature struct {
	Signer crypto.PublicKey `nimona:"signer:s"`
	Alg    string           `nimona:"alg:s"`
	X      []byte           `nimona:"x:d"`
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
	return newSignature(k, m)
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
		m, ok := v.(chore.Map)
		if !ok {
			return true
		}
		if !isObject(m) {
			return true
		}
		meta := &Metadata{}
		mmeta, ok := m["@metadata"].(chore.Map)
		if !ok {
			return true
		}
		err := unmarshalMap(chore.MapHint, mmeta, reflect.ValueOf(meta))
		if err != nil {
			return true
		}
		if !meta.Signature.IsEmpty() {
			return true
		}
		if meta.Owner == nil {
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
		msig, err := sig.MarshalMap()
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
	o.Metadata.Signature = s
	return signErr
}

func isObject(m chore.Map) bool {
	if _, ok := m["@type"]; ok {
		return true
	}
	return false
}

func newSignature(
	k crypto.PrivateKey,
	m chore.Map,
) (Signature, error) {
	h, err := m.Hash().Bytes()
	if err != nil {
		return Signature{}, err
	}
	x := k.Sign(h)
	s := Signature{
		Signer: k.PublicKey(),
		Alg:    AlgorithmObjectHash,
		X:      x,
	}
	return s, nil
}
