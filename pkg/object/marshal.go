package object

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/object/hint"
	"nimona.io/pkg/object/value"
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

func marshalAny(h hint.Hint, v reflect.Value) (value.Value, error) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch h {
	case hint.String:
		if v.Kind() == reflect.String {
			// TODO only for omitempty
			if v.String() == "" {
				return nil, nil
			}
			return value.String(v.String()), nil
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
			return value.String(s), nil
		}
	case hint.CID:
		if v.Kind() == reflect.String {
			// TODO only for omitempty
			if v.String() == "" {
				return nil, nil
			}
			return value.CID(v.String()), nil
		}
	case hint.Bool:
		if v.Kind() == reflect.Bool {
			return value.Bool(v.Bool()), nil
		}
	case hint.Map:
		if v.IsZero() {
			return nil, nil
		}
		if o, isObjPtr := v.Interface().(*Object); isObjPtr {
			if v.IsZero() {
				return nil, nil
			}
			o, err := Marshal(o)
			if err != nil {
				return nil, err
			}
			return o.MarshalMap()
		}
		if o, isObj := v.Interface().(Object); isObj {
			if v.IsZero() {
				return nil, nil
			}
			o, err := Marshal(o)
			if err != nil {
				return nil, err
			}
			return o.MarshalMap()
		}
		if ov, isObj := v.Interface().(ObjectMashaller); isObj {
			if v.IsZero() {
				return nil, nil
			}
			o, err := ov.MarshalObject()
			if err != nil {
				return nil, err
			}
			return o.MarshalMap()
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
	case hint.Float:
		switch v.Kind() {
		case reflect.Float32,
			reflect.Float64:
			return value.Float(v.Float()), nil
		}
	case hint.Int:
		switch v.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			return value.Int(v.Int()), nil
		}
	case hint.Uint:
		switch v.Kind() {
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			return value.Uint(v.Uint()), nil
		}
	case hint.Data:
		m, ok := v.Interface().(ByteMashaller)
		if ok {
			s, err := m.MarshalBytes()
			if err != nil {
				return nil, err
			}
			return value.Data(s), nil
		}
		b, ok := v.Interface().([]byte)
		if ok {
			return value.Data(b), nil
		}
		s, ok := v.Interface().(string)
		if ok {
			b, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return nil, err
			}
			return value.Data(b), nil
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

func marshalStruct(h hint.Hint, v reflect.Value) (value.Map, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.Error("expected struct, got " + v.Kind().String())
	}

	m := value.Map{}
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
		in, ih, err := hint.Extract(ig)
		if err != nil {
			// if there is hint in the key, we check if the value is a primitive
			if ivv, ok := iv.Interface().(value.Value); ok {
				in = ig
				ih = ivv.Hint()
			} else {
				return nil, err
			}
		}
		v, err := marshalAny(ih, iv)
		if err != nil {
			return nil, err
		}
		if v != nil {
			m[in] = v
		}
	}

	// TODO only for omitempty
	if len(m) == 0 {
		return nil, nil
	}

	return m, nil
}

func marshalMap(h hint.Hint, v reflect.Value) (value.Map, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Map {
		return nil, errors.Error("expected map, got " + v.Kind().String())
	}

	if v.Len() == 0 {
		return nil, nil
	}

	m := value.Map{}

	for _, ik := range v.MapKeys() {
		iv := v.MapIndex(ik)
		if ik.Kind() != reflect.String {
			return nil, errors.Error(
				"expected string key, got " + ik.Kind().String(),
			)
		}
		ig := ik.String()
		in, ih, err := hint.Extract(ig)
		if err != nil {
			// if there is hint in the key, we check if the value is a primitive
			if ivv, ok := iv.Interface().(value.Value); ok {
				in = ig
				ih = ivv.Hint()
			} else {
				return nil, err
			}
		}
		v, err := marshalAny(ih, iv)
		if err != nil {
			return nil, err
		}
		m[in] = v
	}

	// TODO only for omitempty
	if len(m) == 0 {
		return nil, nil
	}

	return m, nil
}

func marshalArray(h hint.Hint, v reflect.Value) (value.Value, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return nil, errors.Error("expected slice, got " + v.Kind().String())
	}

	var a value.ArrayValue
	var ah hint.Hint
	switch h {
	case hint.BoolArray:
		a = value.BoolArray{}
		ah = hint.Bool
	case hint.DataArray:
		a = value.DataArray{}
		ah = hint.Data
	case hint.FloatArray:
		a = value.FloatArray{}
		ah = hint.Float
	case hint.IntArray:
		a = value.IntArray{}
		ah = hint.Int
	case hint.MapArray:
		a = value.MapArray{}
		ah = hint.Map
	case hint.StringArray:
		a = value.StringArray{}
		ah = hint.String
	case hint.UintArray:
		a = value.UintArray{}
		ah = hint.Uint
	case hint.CIDArray:
		a = value.CIDArray{}
		ah = hint.CID
	default:
		return nil, errors.Error("unknown array hint")
	}

	for i := 0; i < v.Len(); i++ {
		iv := v.Index(i)
		v, err := marshalAny(ah, iv)
		if err != nil {
			return nil, err
		}

		if v == nil {
			continue
		}

		switch ah {
		case hint.Bool:
			a = append(a.(value.BoolArray), v.(value.Bool))
		case hint.Data:
			a = append(a.(value.DataArray), v.(value.Data))
		case hint.Float:
			a = append(a.(value.FloatArray), v.(value.Float))
		case hint.Int:
			a = append(a.(value.IntArray), v.(value.Int))
		case hint.Map:
			a = append(a.(value.MapArray), v.(value.Map))
		case hint.String:
			a = append(a.(value.StringArray), v.(value.String))
		case hint.Uint:
			a = append(a.(value.UintArray), v.(value.Uint))
		case hint.CID:
			a = append(a.(value.CIDArray), v.(value.CID))
		default:
			return nil, errors.Error("unknown array element hint " + ah)
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
