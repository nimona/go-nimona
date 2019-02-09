package object

import (
	"fmt"
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
	HintMap       TypeHint = "o"
	HintNil       TypeHint = "n"
	HintString    TypeHint = "s"
	HintUint      TypeHint = "u"
)

// DeduceTypeHint returns a TypeHint from a given value
func DeduceTypeHint(o interface{}) TypeHint {
	t := reflect.TypeOf(o)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		if t.Elem() == reflect.TypeOf(byte(0)) {
			return HintData
		}
		// TODO(geoah) add support for A<A<*>>
		// eg {"foo:A<A<i>>:" [["1", "2"], ["3", "4"]]} should be A<A<i>>
		sv := reflect.New(t.Elem()).Elem().Interface()
		if sv == nil {
			oo := o.([]interface{})
			if len(oo) > 0 {
				sv = oo[0]
			}
		}
		if sv != nil {
			subType := DeduceTypeHint(sv)
			return HintArray + "<" + subType + ">"
		}
		return HintArray + "<?>" // TODO(geoah) should this return "" or panic maybe?

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

	fmt.Println("___________ COULD NOT DEDUCE", o) // TODO LOG

	return HintUndefined
}
