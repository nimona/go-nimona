package blocks

import (
	"reflect"

	"nimona.io/go/crypto"
)

type Typed interface {
	GetType() string
	GetSignature() *crypto.Signature
	SetSignature(*crypto.Signature)
	GetAnnotations() map[string]interface{}
	SetAnnotations(map[string]interface{})
}

var (
	typedType = reflect.TypeOf(new(Typed)).Elem()
)
