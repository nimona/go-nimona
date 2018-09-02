package blocks

import "reflect"

const DefaultTagName = "nimona"

// Marshaler is the interface implemented by types that
// can marshal themselves into valid Block.
type Marshaler interface {
	MarshalBlock() (string, error)
}

type Unmarshaler interface {
	UnmarshalBlock(string) error
}

var (
	marshalerType   = reflect.TypeOf(new(Marshaler)).Elem()
	unmarshalerType = reflect.TypeOf(new(Unmarshaler)).Elem()
	// textMarshalerType = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()
)
