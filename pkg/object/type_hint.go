package object

import (
	"reflect"
)

type (
	// TypeHint are the hints of a member's type
	TypeHint string
)

// String implements the Stringer interface
func (t TypeHint) String() string {
	return string(t)
}

const (
	HintUndefined TypeHint = ""
	HintObject    TypeHint = "o"
	HintArray     TypeHint = "a"
	HintBool      TypeHint = "b"
	HintData      TypeHint = "d"
	HintFloat     TypeHint = "f"
	HintInt       TypeHint = "i"
	HintNil       TypeHint = "n"
	HintString    TypeHint = "s"
	HintUint      TypeHint = "u"
)

var (
	hints = map[string]TypeHint{
		"":  HintUndefined,
		"o": HintObject,
		"a": HintArray,
		"b": HintBool,
		"d": HintData,
		"f": HintFloat,
		"i": HintInt,
		"n": HintNil,
		"s": HintString,
		"u": HintUint,
	}
)

// GetTypeHint returns a TypeHint from a string
func GetTypeHint(t string) TypeHint {
	if t, ok := hints[t]; ok {
		return t
	}
	return HintUndefined
}

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
		return HintObject

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
