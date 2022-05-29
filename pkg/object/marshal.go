package object

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
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

	metaOwner, err := marshalPickSpecial(v, "@metadata.owner:s")
	if err != nil {
		return nil, err
	}
	if v, ok := metaOwner.(peer.ID); ok {
		o.Metadata.Owner = v
	}

	metaTimestamp, err := marshalPickSpecial(v, "@metadata.timestamp:s")
	if err != nil {
		return nil, err
	}
	if v, ok := metaTimestamp.(string); ok {
		o.Metadata.Timestamp = v
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
		o.Context = tilde.Digest(v)
	}

	delete(m, "@metadata")
	delete(m, "@metadata.owner")
	delete(m, "@metadata.timestamp")

	if t, ok := m["@type"]; ok {
		if tt, ok := t.(tilde.String); ok {
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
		if ig == k && iv.CanInterface() {
			return iv.Interface(), nil
		}
	}

	return nil, nil
}

func marshalAny(
	h tilde.Hint,
	tagOptions map[string]string,
	v reflect.Value,
) (tilde.Value, error) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch h {
	case tilde.StringHint:
		if v.Kind() == reflect.String {
			// TODO only for omitempty
			if v.String() == "" {
				return nil, nil
			}
			return tilde.String(v.String()), nil
		}
		m, ok := v.Interface().(StringMashaller)
		if ok {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					return nil, nil
				}
			}
			s, err := m.MarshalString()
			if err != nil {
				return nil, err
			}
			// TODO only for omitempty
			if s == "" {
				return nil, nil
			}
			return tilde.String(s), nil
		}
	case tilde.DigestHint:
		if b, ok := v.Interface().([]byte); ok {
			// TODO only for omitempty
			if b == nil {
				return nil, nil
			}
			return tilde.Digest(b), nil
		}
		if b, ok := v.Interface().(tilde.Digest); ok {
			// TODO only for omitempty
			if b.IsEmpty() {
				return nil, nil
			}
			return b, nil
		}
	case tilde.BoolHint:
		if v.Kind() == reflect.Bool {
			return tilde.Bool(v.Bool()), nil
		}
	case tilde.MapHint:
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
	case tilde.FloatHint:
		switch v.Kind() {
		case reflect.Float32,
			reflect.Float64:
			if _, ok := tagOptions["omitzero"]; ok && v.IsZero() {
				return nil, nil
			}
			return tilde.Float(v.Float()), nil
		}
	case tilde.IntHint:
		switch v.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			if _, ok := tagOptions["omitzero"]; ok && v.IsZero() {
				return nil, nil
			}
			return tilde.Int(v.Int()), nil
		}
	case tilde.UintHint:
		switch v.Kind() {
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			if _, ok := tagOptions["omitzero"]; ok && v.IsZero() {
				return nil, nil
			}
			return tilde.Uint(v.Uint()), nil
		}
	case tilde.DataHint:
		m, ok := v.Interface().(ByteMashaller)
		if ok {
			s, err := m.MarshalBytes()
			if err != nil {
				return nil, err
			}
			return tilde.Data(s), nil
		}
		b, ok := v.Interface().([]byte)
		if ok {
			return tilde.Data(b), nil
		}
		s, ok := v.Interface().(string)
		if ok {
			b, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return nil, err
			}
			return tilde.Data(b), nil
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

func marshalStruct(h tilde.Hint, v reflect.Value) (tilde.Map, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.Error("expected struct, got " + v.Kind().String())
	}

	m := tilde.Map{}
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
			// we check can interface to allow for unexported metadata such
			// as how the signature works
			if iv.CanInterface() {
				if _, ok := iv.Interface().(Metadata); ok {
					imm, err := marshalStruct(tilde.MapHint, iv)
					if err != nil {
						return nil, err
					}
					m["@metadata"] = imm
				}
			}
			if t, ok := igKvs["context"]; ok {
				m["@context"] = tilde.Digest(t)
			}
			if t, ok := igKvs["type"]; ok {
				m["@type"] = tilde.String(t)
			}
			continue
		}
		// handle special cases where we don't have full `@metadata`, but type
		// has been added to an individual metadata field, ie `@metadata.owner`
		// TODO: should we make sure only one field has the type set?
		if strings.HasPrefix(ig, "@metadata.") {
			if t, ok := igKvs["type"]; ok {
				m["@type"] = tilde.String(t)
			}
		}
		// look for the field's hint
		in, ih, err := tilde.ExtractHint(ig)
		if err != nil {
			// if there is hint in the key, we check if the value is a primitive
			if ivv, ok := iv.Interface().(tilde.Value); ok {
				in = ig
				ih = ivv.Hint()
			} else {
				return nil, err
			}
		}
		v, err := marshalAny(ih, igKvs, iv)
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

func marshalMap(h tilde.Hint, v reflect.Value) (tilde.Map, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Map {
		return nil, errors.Error("expected map, got " + v.Kind().String())
	}

	if v.Len() == 0 {
		return nil, nil
	}

	m := tilde.Map{}

	for _, ik := range v.MapKeys() {
		iv := v.MapIndex(ik)
		if ik.Kind() != reflect.String {
			return nil, errors.Error(
				"expected string key, got " + ik.Kind().String(),
			)
		}
		ig := ik.String()
		in, ih, err := tilde.ExtractHint(ig)
		if err != nil {
			// if there is hint in the key, we check if the value is a primitive
			if ivv, ok := iv.Interface().(tilde.Value); ok {
				in = ig
				ih = ivv.Hint()
			} else {
				return nil, err
			}
		}
		v, err := marshalAny(ih, nil, iv)
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

func marshalArray(h tilde.Hint, v reflect.Value) (tilde.Value, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return nil, errors.Error("expected slice, got " + v.Kind().String())
	}

	var a tilde.ArrayValue
	var ah tilde.Hint
	switch h {
	case tilde.BoolArrayHint:
		a = tilde.BoolArray{}
		ah = tilde.BoolHint
	case tilde.DataArrayHint:
		a = tilde.DataArray{}
		ah = tilde.DataHint
	case tilde.FloatArrayHint:
		a = tilde.FloatArray{}
		ah = tilde.FloatHint
	case tilde.IntArrayHint:
		a = tilde.IntArray{}
		ah = tilde.IntHint
	case tilde.MapArrayHint:
		a = tilde.MapArray{}
		ah = tilde.MapHint
	case tilde.StringArrayHint:
		a = tilde.StringArray{}
		ah = tilde.StringHint
	case tilde.UintArrayHint:
		a = tilde.UintArray{}
		ah = tilde.UintHint
	case tilde.DigestArrayHint:
		a = tilde.DigestArray{}
		ah = tilde.DigestHint
	default:
		return nil, errors.Error("unknown array hint")
	}

	for i := 0; i < v.Len(); i++ {
		iv := v.Index(i)
		v, err := marshalAny(ah, nil, iv)
		if err != nil {
			return nil, err
		}

		if v == nil {
			continue
		}

		switch ah {
		case tilde.BoolHint:
			a = append(a.(tilde.BoolArray), v.(tilde.Bool))
		case tilde.DataHint:
			a = append(a.(tilde.DataArray), v.(tilde.Data))
		case tilde.FloatHint:
			a = append(a.(tilde.FloatArray), v.(tilde.Float))
		case tilde.IntHint:
			a = append(a.(tilde.IntArray), v.(tilde.Int))
		case tilde.MapHint:
			a = append(a.(tilde.MapArray), v.(tilde.Map))
		case tilde.StringHint:
			a = append(a.(tilde.StringArray), v.(tilde.String))
		case tilde.UintHint:
			a = append(a.(tilde.UintArray), v.(tilde.Uint))
		case tilde.DigestHint:
			a = append(a.(tilde.DigestArray), v.(tilde.Digest))
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
		switch len(kvParts) {
		case 1:
			kvs[kvParts[0]] = ""
		case 2:
			kvs[kvParts[0]] = kvParts[1]
		default:
			return "", nil, errors.Error("invalid tag options")
		}
	}
	return name, kvs, nil
}
