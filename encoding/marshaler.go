package encoding

import "reflect"

const DefaultTagName = "nimona"

// Marshaler is the interface implemented by types that
// can marshal themselves into valid Block.
type Marshaler interface {
	MarshalBlock() ([]byte, error)
}

var (
	marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
	// textMarshalerType = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()
)
