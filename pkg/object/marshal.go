package object

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"nimona.io/pkg/errors"
)

func MustMarshal(in interface{}) *Object {
	o, err := Marshal(in)
	if err != nil {
		panic(err)
	}
	return o
}

func Marshal(in interface{}) (*Object, error) {
	if o, ok := in.(*Object); ok {
		return o, nil
	}

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

	octx, err := marshalPickSpecial(v, "@context:s")
	if err != nil {
		return nil, err
	}
	if v, ok := octx.(string); ok {
		o.Context = v
	}

	if o.Type == "" {
		tr, ok := in.(Typed)
		if ok {
			o.Type = tr.Type()
		}
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
			continue
		}
		ig, err := getStructTagName(it)
		if err != nil {
			return nil, fmt.Errorf("marshal special: attribute %s, %w", it.Name, err)
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
	switch h {
	case StringHint:
		if v.Kind() == reflect.String {
			// TODO only for omitempty
			if v.String() == "" {
				return nil, nil
			}
			return String(v.String()), nil
		}
		m, ok := v.Interface().(StringMashaller)
		if ok {
			s, err := m.MarshalString()
			if err != nil {
				return nil, err
			}
			// TODO only for omitempty
			if s == "" {
				return nil, nil
			}
			return String(s), nil
		}
	case CIDHint:
		if v.Kind() == reflect.String {
			// TODO only for omitempty
			if v.String() == "" {
				return nil, nil
			}
			return CID(v.String()), nil
		}
	case BoolHint:
		if v.Kind() == reflect.Bool {
			return Bool(v.Bool()), nil
		}
	case MapHint:
		if v.IsZero() {
			return nil, nil
		}
		switch v.Kind() {
		case reflect.Map:
			return marshalMap(h, v)
		case reflect.Ptr:
			if v.IsNil() {
				return nil, nil
			}
			v = v.Elem()
			if !v.IsValid() {
				return nil, nil
			}
			fallthrough
		case reflect.Struct:
			return marshalStruct(h, v)
		}
	case ObjectHint:
		if v.IsZero() {
			return nil, nil
		}
		return Marshal(v.Interface())
	case FloatHint:
		switch v.Kind() {
		case reflect.Float32,
			reflect.Float64:
			return Float(v.Float()), nil
		}
	case IntHint:
		switch v.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			return Int(v.Int()), nil
		}
	case UintHint:
		switch v.Kind() {
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			return Uint(v.Uint()), nil
		}
	case DataHint:
		m, ok := v.Interface().(ByteMashaller)
		if ok {
			s, err := m.MarshalBytes()
			if err != nil {
				return nil, err
			}
			return Data(s), nil
		}
		b, ok := v.Interface().([]byte)
		if ok {
			return Data(b), nil
		}
		s, ok := v.Interface().(string)
		if ok {
			b, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return nil, err
			}
			return Data(b), nil
		}
	}
	if h[0] == 'a' {
		switch v.Kind() {
		case reflect.Array,
			reflect.Slice:
			return marshalArray(h, v)
		}
	}
	return nil, errors.Error("invalid type " + v.Kind().String() +
		" for hint " + string(h))
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
			return nil, fmt.Errorf("marshal: attribute %s, %w", it.Name, err)
		}
		switch ig {
		case "@type:s",
			"@context:s",
			"@metadata:m":
			continue
		}
		in, ih, err := splitHint([]byte(ig))
		if err != nil {
			// if there is hint in the key, we check if the value is a primitive
			if ivv, ok := iv.Interface().(Value); ok {
				in = ig
				ih = ivv.Hint()
			} else {
				return nil, err
			}
		}
		value, err := marshalAny(ih, iv)
		if err != nil {
			return nil, err
		}
		if value != nil {
			m[in] = value
		}
	}

	// TODO only for omitempty
	if len(m) == 0 {
		return nil, nil
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

	if v.Len() == 0 {
		return nil, nil
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
			// if there is hint in the key, we check if the value is a primitive
			if ivv, ok := iv.Interface().(Value); ok {
				in = ig
				ih = ivv.Hint()
			} else {
				return nil, err
			}
		}
		value, err := marshalAny(ih, iv)
		if err != nil {
			return nil, err
		}
		m[in] = value
	}

	// TODO only for omitempty
	if len(m) == 0 {
		return nil, nil
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

		if value == nil {
			continue
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

	// TODO only for omitempty
	if a.Len() == 0 {
		return nil, nil
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
