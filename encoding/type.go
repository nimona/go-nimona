package encoding

import (
	"errors"
	"fmt"
	"reflect"
)

// TypeMap will add hint suffixes on the map's keys depending on the value's type.
// It is based on TJSON with the only difference being that `d` is used for raw
// bytes instead of encoded. And there are no hints for b16, b32, and b64.
func TypeMap(m map[string]interface{}) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	for k, v := range m {
		t := reflect.TypeOf(v)
		h := GetHintFromType(v)
		if h == "" {
			panic(fmt.Sprintf("type: unsupported type k=%s t=%s v=%#v", k, t.String(), v))
		}
		eh := getFullType(k)
		if eh == "" {
			// if key doesn't have a type, add it
			k += ":" + h
		} else {
			// if key already has type, check that it's the same as the one
			// we are expecting it to have
			if h != eh {
				return nil, fmt.Errorf("type: type hinted as %s, but is %s", eh, h)
			}
		}
		switch t.Kind() {
		// TODO(geoah) do we need to panic for structs?
		// TODO(geoah) add A<O> support, check tests
		case reflect.Map:
			m, ok := v.(map[string]interface{})
			if !ok {
				return nil, errors.New("only map[string]interface{} are supported")
			}
			vs, err := TypeMap(m)
			if err != nil {
				return nil, err
			}
			out[k] = vs
		default:
			out[k] = v
		}
	}
	return out, nil
}

func GetHintFromType(o interface{}) string {
	// v := reflect.ValueOf(o)
	t := reflect.TypeOf(o)
	// if o == nil {
	// 	fmt.Println("asdfasdf")
	// }
	// fmt.Println("---", o)
	// fmt.Println("---", v)
	// fmt.Println("---t", t.Kind())
	// fmt.Println("---t", v.Kind())
	switch t.Kind() {
	case reflect.Invalid:
		// TODO(geoah) erm when would this happen?
		panic("typed cannot handle invalid")
	case reflect.Slice, reflect.Array:
		if t.Elem() == reflect.TypeOf(byte(0)) {
			return HintBytes
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
			// fmt.Println("ppppp", sv)
			subType := GetHintFromType(sv)
			return HintArray + "<" + subType + ">"
		}
		return HintArray + "<?>" // TODO(geoah) should this return "" or panic maybe?
	case reflect.String:
		return HintString
	case reflect.Map, reflect.Struct: // TODO(geoah) do we need to panic for structs?
		return HintMap
	case reflect.Float32, reflect.Float64:
		return HintFloat
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return HintInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return HintUint
	// case reflect.TypeOf(&big.Int{}).Kind():
	// 	return HintInt
	case reflect.Bool:
		return HintBool
		// case reflect.Ptr:
		// 	return getTypeHint(v.Interface())
	}
	return ""
	// panic("hint: unsupported type " + v.String() + " -- " + fmt.Sprintf("%#v", v))
}

// func getTypeHint(v reflect.Type) string {
// 	switch v.Kind() {
// 	case reflect.Invalid:
// 		// TODO(geoah) erm when would this happen?
// 		panic("typed cannot handle invalid")
// 	case reflect.Slice, reflect.Array:
// 		if v.Elem() == reflect.TypeOf(byte(0)) {
// 			return hintBytes
// 		}
// 		// TODO(geoah) add support for A<A<*>>
// 		// eg {"foo:A<A<i>>:" [["1", "2"], ["3", "4"]]} should be A<A<i>>
// 		subType := getTypeHint(v.Elem())
// 		return hintArray + "<" + subType + ">"
// 	case reflect.String:
// 		return hintString
// 	case reflect.Map, reflect.Struct: // TODO(geoah) do we need to panic for structs?
// 		return hintMap
// 	case reflect.Float32, reflect.Float64:
// 		return hintFloat
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		return hintInt
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		return hintUint
// 	// case reflect.TypeOf(&big.Int{}).Kind():
// 	// 	return hintInt
// 	case reflect.Bool:
// 		return hintBool
// 		// case reflect.Ptr:
// 		// 	return getTypeHint(v.Elem())
// 	}
// 	return ""
// 	// panic("hint: unsupported type " + v.String() + " -- " + fmt.Sprintf("%#v", v))
// }
