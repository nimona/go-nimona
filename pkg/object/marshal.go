package object

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"nimona.io/pkg/chore"
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

	if v.Kind() != reflect.Ptr {
		panic("marshal currently doesn't support non pointers")
	}

	m, err := marshalStruct(":m", v)
	if err != nil {
		return nil, err
	}
	o := &Object{
		Data: m,
	}

	// TODO consider rewritting the pick specials

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

	delete(m, "@metadata")

	if t, ok := m["@type"]; ok {
		if tt, ok := t.(chore.String); ok {
			o.Type = string(tt)
			delete(m, "@type")
		}
	}

	if o.Type == "" {
		tr, ok := in.(Typer)
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
		ig, _, err := getStructTagName(it)
		if err != nil {
			return nil, fmt.Errorf("marshal special: attribute %s, %w", it.Name, err)
		}
		if ig == k {
			return iv.Interface(), nil
		}
	}

	return nil, nil
}

func marshalAny(h chore.Hint, v reflect.Value) (chore.Value, error) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch h {
	case chore.StringHint:
		if v.Kind() == reflect.String {
			// TODO only for omitempty
			if v.String() == "" {
				return nil, nil
			}
			return chore.String(v.String()), nil
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
			return chore.String(s), nil
		}
	case chore.HashHint:
		if b, ok := v.Interface().([]byte); ok {
			// TODO only for omitempty
			if b == nil {
				return nil, nil
			}
			return chore.Hash(b), nil
		}
		if b, ok := v.Interface().(chore.Hash); ok {
			// TODO only for omitempty
			if b.IsEmpty() {
				return nil, nil
			}
			return b, nil
		}
	case chore.BoolHint:
		if v.Kind() == reflect.Bool {
			return chore.Bool(v.Bool()), nil
		}
	case chore.MapHint:
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
		if ov, isObj := v.Interface().(Typer); isObj {
			if v.IsZero() {
				return nil, nil
			}
			o, err := Marshal(ov)
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
	case chore.FloatHint:
		switch v.Kind() {
		case reflect.Float32,
			reflect.Float64:
			return chore.Float(v.Float()), nil
		}
	case chore.IntHint:
		switch v.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			return chore.Int(v.Int()), nil
		}
	case chore.UintHint:
		switch v.Kind() {
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			return chore.Uint(v.Uint()), nil
		}
	case chore.DataHint:
		m, ok := v.Interface().(ByteMashaller)
		if ok {
			s, err := m.MarshalBytes()
			if err != nil {
				return nil, err
			}
			return chore.Data(s), nil
		}
		b, ok := v.Interface().([]byte)
		if ok {
			return chore.Data(b), nil
		}
		s, ok := v.Interface().(string)
		if ok {
			b, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return nil, err
			}
			return chore.Data(b), nil
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

func marshalStruct(h chore.Hint, v reflect.Value) (chore.Map, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.Error("expected struct, got " + v.Kind().String())
	}

	m := chore.Map{}
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
		ig, igKvs, err := getStructTagName(it)
		if err != nil {
			return nil, fmt.Errorf("marshal: attribute %s, %w", it.Name, err)
		}
		switch ig {
		case "@type:s",
			"@context:s":
			continue
		case "@metadata:m":
			if _, ok := iv.Interface().(Metadata); ok {
				imm, err := marshalStruct(chore.MapHint, iv)
				if err != nil {
					return nil, err
				}
				m["@metadata"] = imm
			}
			if t, ok := igKvs["type"]; ok {
				m["@type"] = chore.String(t)
			}
			continue
		}
		in, ih, err := chore.ExtractHint(ig)
		if err != nil {
			// if there is hint in the key, we check if the value is a primitive
			if ivv, ok := iv.Interface().(chore.Value); ok {
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

func marshalMap(h chore.Hint, v reflect.Value) (chore.Map, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Map {
		return nil, errors.Error("expected map, got " + v.Kind().String())
	}

	if v.Len() == 0 {
		return nil, nil
	}

	m := chore.Map{}

	for _, ik := range v.MapKeys() {
		iv := v.MapIndex(ik)
		if ik.Kind() != reflect.String {
			return nil, errors.Error(
				"expected string key, got " + ik.Kind().String(),
			)
		}
		ig := ik.String()
		in, ih, err := chore.ExtractHint(ig)
		if err != nil {
			// if there is hint in the key, we check if the value is a primitive
			if ivv, ok := iv.Interface().(chore.Value); ok {
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

func marshalArray(h chore.Hint, v reflect.Value) (chore.Value, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return nil, errors.Error("expected slice, got " + v.Kind().String())
	}

	var a chore.ArrayValue
	var ah chore.Hint
	switch h {
	case chore.BoolArrayHint:
		a = chore.BoolArray{}
		ah = chore.BoolHint
	case chore.DataArrayHint:
		a = chore.DataArray{}
		ah = chore.DataHint
	case chore.FloatArrayHint:
		a = chore.FloatArray{}
		ah = chore.FloatHint
	case chore.IntArrayHint:
		a = chore.IntArray{}
		ah = chore.IntHint
	case chore.MapArrayHint:
		a = chore.MapArray{}
		ah = chore.MapHint
	case chore.StringArrayHint:
		a = chore.StringArray{}
		ah = chore.StringHint
	case chore.UintArrayHint:
		a = chore.UintArray{}
		ah = chore.UintHint
	case chore.HashArrayHint:
		a = chore.HashArray{}
		ah = chore.HashHint
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
		case chore.BoolHint:
			a = append(a.(chore.BoolArray), v.(chore.Bool))
		case chore.DataHint:
			a = append(a.(chore.DataArray), v.(chore.Data))
		case chore.FloatHint:
			a = append(a.(chore.FloatArray), v.(chore.Float))
		case chore.IntHint:
			a = append(a.(chore.IntArray), v.(chore.Int))
		case chore.MapHint:
			a = append(a.(chore.MapArray), v.(chore.Map))
		case chore.StringHint:
			a = append(a.(chore.StringArray), v.(chore.String))
		case chore.UintHint:
			a = append(a.(chore.UintArray), v.(chore.Uint))
		case chore.HashHint:
			a = append(a.(chore.HashArray), v.(chore.Hash))
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

// getStructTagName splits the tag into a name, and a number of key-value pars
// - `nimona:"@metadata:m"
// - `nimona:"@metadata:m,type=foo"
// - `nimona:"@metadata:m,type=foo,x=y"
func getStructTagName(f reflect.StructField) (
	name string,
	kvs map[string]string,
	err error,
) {
	tag := strings.TrimPrefix(string(f.Tag), "nimona:")
	tag = strings.Trim(tag, `"`)
	tagParts := strings.Split(tag, ",")
	name = tagParts[0]
	if name == "" {
		return "", nil, errors.Error("tag name cannot be empty")
	}
	if len(tagParts) == 1 {
		return name, nil, nil
	}
	kvs = map[string]string{}
	for _, kv := range tagParts[1:] {
		kvParts := strings.Split(kv, "=")
		if len(kvParts) != 2 {
			return "", nil, errors.Error("invalid key-value parts")
		}
		kvs[kvParts[0]] = kvParts[1]
	}
	return name, kvs, nil
}
