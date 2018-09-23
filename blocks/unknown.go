package blocks

import (
	"reflect"

	"nimona.io/go/crypto"
)

var (
	unknownType = reflect.TypeOf(&Unknown{})
)

type Unknown struct {
	Type        string                 `json:"-"`
	Payload     map[string]interface{} `json:"payload,payload"`
	Annotations map[string]interface{} `json:"-"`
	Signature   *crypto.Signature      `json:"-"`
}

func (r *Unknown) GetType() string {
	return r.Type
}

func (r *Unknown) GetSignature() *crypto.Signature {
	return r.Signature
}

func (r *Unknown) SetSignature(s *crypto.Signature) {
	r.Signature = s
}

func (r *Unknown) GetAnnotations() map[string]interface{} {
	return r.Annotations
}

func (r *Unknown) SetAnnotations(a map[string]interface{}) {
	r.Annotations = a
}
