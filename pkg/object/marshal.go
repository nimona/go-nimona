package object

import (
	"fmt"
	"reflect"
	"strings"

	"nimona.io/pkg/errors"
)

func Marshal(in interface{}) (*Object, error) {
	v := reflect.ValueOf(in)
	m, err := marshalStruct(":m", v)
	if err != nil {
		return nil, err
	}
	o := &Object{
		Data: m,
	}

	meta, err := marshalPickSpecial(v, "@metadata:m")
	if err != nil {
		return nil, err
	}
	if v, ok := meta.(Metadata); ok {
		o.Metadata = v
	}

	otype, err := marshalPickSpecial(v, "@type:s")
	if err != nil {
		return nil, err
	}
	if v, ok := otype.(string); ok {
		o.Type = v
	}

	return o, nil
}

func marshalPickSpecial(v reflect.Value, k string) (interface{}, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.Error("expected struct, got " + v.Kind().String())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		iv := v.Field(i)
		it := t.Field(i)
		if it.Anonymous {
			s, err := marshalPickSpecial(iv, k)
			if err != nil {
				return nil, err
			}
			if s != nil {
				return s, nil
			}
		}
		ig, err := getStructTagName(it)
		if err != nil {
			return nil, fmt.Errorf("attribute %s, %w", it.Name, err)
		}
		if ig == k {
			return iv.Interface(), nil
		}
	}

	return nil, nil
}

func marshalAny(h Hint, v reflect.Value) (Value, error) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return String(v.String()), nil
	case reflect.Bool:
		return Bool(v.Bool()), nil
	case reflect.Map:
		return marshalMap(h, v)
	case reflect.Struct:
		return marshalStruct(h, v)
	case reflect.Array,
		reflect.Slice:
		return marshalArray(h, v)
	case reflect.Float32,
		reflect.Float64:
		return Float(v.Float()), nil
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return Int(v.Int()), nil
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return Uint(v.Uint()), nil
	}
	return nil, errors.Error("unknown type " + v.Kind().String())
}

func marshalStruct(h Hint, v reflect.Value) (Map, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.Error("expected struct, got " + v.Kind().String())
	}

	m := Map{}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		iv := v.Field(i)
		it := t.Field(i)
		if it.Anonymous {
			em, err := marshalStruct(h, iv)
			if err != nil {
				return nil, err
			}
			for k, v := range em {
				m[k] = v
			}
			continue
		}
		ig, err := getStructTagName(it)
		if err != nil {
			return nil, fmt.Errorf("attribute %s, %w", it.Name, err)
		}
		switch ig {
		case "@type:s",
			"@metadata:m":
			continue
		}
		in, ih, err := splitHint([]byte(ig))
		if err != nil {
			return nil, err
		}
		value, err := marshalAny(ih, iv)
		if err != nil {
			return nil, err
		}
		m[in] = value
	}

	return m, nil
}

func marshalMap(h Hint, v reflect.Value) (Map, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Map {
		return nil, errors.Error("expected map, got " + v.Kind().String())
	}

	m := Map{}

	for _, ik := range v.MapKeys() {
		iv := v.MapIndex(ik)
		if ik.Kind() != reflect.String {
			return nil, errors.Error(
				"expected string key, got " + ik.Kind().String(),
			)
		}
		ig := ik.String()
		in, ih, err := splitHint([]byte(ig))
		if err != nil {
			return nil, err
		}
		value, err := marshalAny(ih, iv)
		if err != nil {
			return nil, err
		}
		m[in] = value
	}

	return m, nil
}

func marshalArray(h Hint, v reflect.Value) (Value, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return nil, errors.Error("expected slice, got " + v.Kind().String())
	}

	var a ArrayValue
	var ah Hint
	switch h {
	case BoolArrayHint:
		a = BoolArray{}
		ah = BoolHint
	case DataArrayHint:
		a = DataArray{}
		ah = DataHint
	case FloatArrayHint:
		a = FloatArray{}
		ah = FloatHint
	case IntArrayHint:
		a = IntArray{}
		ah = IntHint
	case MapArrayHint:
		a = MapArray{}
		ah = MapHint
	case ObjectArrayHint:
		a = ObjectArray{}
		ah = ObjectHint
	case StringArrayHint:
		a = StringArray{}
		ah = StringHint
	case UintArrayHint:
		a = UintArray{}
		ah = UintHint
	case CIDArrayHint:
		a = CIDArray{}
		ah = CIDHint
	default:
		return nil, errors.Error("unknown array hint")
	}

	for i := 0; i < v.Len(); i++ {
		iv := v.Index(i)
		value, err := marshalAny(ah, iv)
		if err != nil {
			return nil, err
		}

		switch ah {
		case BoolHint:
			a = append(a.(BoolArray), value.(Bool))
		case DataHint:
			a = append(a.(DataArray), value.(Data))
		case FloatHint:
			a = append(a.(FloatArray), value.(Float))
		case IntHint:
			a = append(a.(IntArray), value.(Int))
		case MapHint:
			a = append(a.(MapArray), value.(Map))
		case ObjectHint:
			a = append(a.(ObjectArray), value.(*Object))
		case StringHint:
			a = append(a.(StringArray), value.(String))
		case UintHint:
			a = append(a.(UintArray), value.(Uint))
		case CIDHint:
			a = append(a.(CIDArray), value.(CID))
		default:
			return nil, errors.Error("unknown array element hint")
		}
	}

	return a, nil
}

func getStructTagName(f reflect.StructField) (string, error) {
	v := strings.Replace(string(f.Tag), "nimona:", "", 1)
	if v == "" {
		return "", errors.Error("tag cannot be empty")
	}
	return strings.Trim(v, `"`), nil
}
