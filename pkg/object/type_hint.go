package object

import (
	"reflect"
)

type (
	// TypeHint are the hints of a member's type
	TypeHint string
	// TODO v is only used for arrays, we should find a better way to do this.
	typeHints []TypeHint
)

// String implements the Stringer interface
func (t TypeHint) String() string {
	return string(t)
}

// String implements the Stringer interface
func (ts typeHints) String() string {
	s := ""
	for _, t := range ts {
		s += t.String()
	}
	return s
}

// TypeHint for a composite hint.
func (ts typeHints) TypeHint() TypeHint {
	s := ""
	for _, t := range ts {
		s += t.String()
	}
	return TypeHint(s)
}

const (
	HintUndefined TypeHint = ""
	HintMap       TypeHint = "m"
	HintArray     TypeHint = "a"
	HintBool      TypeHint = "b"
	HintData      TypeHint = "d"
	HintFloat     TypeHint = "f"
	HintInt       TypeHint = "i"
	HintNil       TypeHint = "n"
	HintString    TypeHint = "s"
	HintRef       TypeHint = "r"
	HintUint      TypeHint = "u"
)

// DeduceTypeHint returns a TypeHint from a given value
func DeduceTypeHint(o interface{}) TypeHint {
	if o == nil {
		return HintUndefined
	}

	t := reflect.TypeOf(o)
	if t == nil {
		return HintUndefined
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		if t.Elem() == reflect.TypeOf(byte(0)) {
			return HintData
		}
		// TODO(geoah) add support for aa*
		// eg {"foo:aai:" [["1", "2"], ["3", "4"]]} should be aai
		sv := reflect.New(t.Elem()).Elem().Interface()
		if sv == nil {
			oo := o.([]interface{})
			if len(oo) > 0 {
				sv = oo[0]
			}
		}
		if sv != nil {
			subType := DeduceTypeHint(sv)
			return HintArray + subType
		}
		return HintArray + "?" // TODO(geoah) should this return "" or panic maybe?

	case reflect.String:
		return HintString

	case reflect.Map,
		reflect.Struct:
		return HintMap

	case reflect.Float32, reflect.Float64:
		return HintFloat

	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return HintInt

	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return HintUint

	case reflect.Bool:
		return HintBool
	}

	return HintUndefined
}
