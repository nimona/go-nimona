package object

import (
	"fmt"
	"reflect"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/object/hint"
	"nimona.io/pkg/object/value"
)

// Unmarshal an object into a tagged struct
func Unmarshal(o *Object, out interface{}) error {
	if out == nil || o == nil {
		return nil
	}
	v := reflect.ValueOf(out)
	err := unmarshalSpecials(o, v)
	if err != nil {
		return err
	}
	return unmarshalMap(hint.Map, o.Data, v)
}

func unmarshalSpecials(o *Object, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.Error("unmarshal special: expected struct, got " +
			v.Kind().String())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		iv := v.Field(i)
		it := t.Field(i)
		if it.Anonymous {
			return unmarshalSpecials(o, iv)
		}
		ig, err := getStructTagName(it)
		if err != nil {
			continue
		}
		switch ig {
		case "@type:s":
			iv.Set(reflect.ValueOf(o.Type))
		case "@context:s":
			iv.Set(reflect.ValueOf(o.Context))
		case "@metadata:m":
			iv.Set(reflect.ValueOf(o.Metadata))
		}
	}
	return nil
}

func unmarshalMap(h hint.Hint, m value.Map, target reflect.Value) error {
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	if _, hasType := m["@type"]; hasType {
		o := &Object{}
		err := o.UnmarshalMap(m)
		if err != nil {
			return err
		}
		var ev reflect.Value
		if target.Kind() == reflect.Ptr {
			ev = reflect.New(target.Type().Elem())
		} else {
			ev = reflect.New(target.Type())
		}
		// if the target is an object simply set it
		if _, isObjPtr := ev.Interface().(*Object); isObjPtr {
			target.Set(reflect.ValueOf(*o))
			return nil
		}
		if _, isObj := ev.Interface().(Object); isObj {
			target.Set(reflect.ValueOf(*o))
			return nil
		}
		// else we should try to unmarshal onto it
		err = Unmarshal(o, ev.Interface())
		if err != nil {
			return err
		}
		if target.Kind() == reflect.Ptr {
			target.Set(ev)
		} else {
			target.Set(ev.Elem())
		}
	}

	switch target.Kind() {
	case reflect.Struct:
		return unmarshalMapToStruct(h, m, target)
	case reflect.Map:
		return unmarshalMapToMap(h, m, target)
	}

	return errors.Error("expected map or struct, got " + target.Kind().String())
}

func unmarshalMapToStruct(h hint.Hint, m value.Map, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors.Error("unmarshal: expected struct, got " + v.Kind().String())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		iv := v.Field(i)
		it := t.Field(i)
		if it.Anonymous {
			return unmarshalMapToStruct(h, m, iv)
		}
		ig, err := getStructTagName(it)
		if err != nil {
			return fmt.Errorf("unmarshal map: attribute %s, %w", it.Name, err)
		}
		switch ig {
		case "@type:s",
			"@context:s",
			"@metadata:m":
			// TODO special
			continue
		}
		in, ih, err := hint.Extract(ig)
		if err != nil {
			return err
		}
		val, ok := m[in]
		if !ok {
			continue
		}
		if err := unmarshalAny(ih, val, iv); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalMapToMap(h hint.Hint, m value.Map, v reflect.Value) error {
	if v.Kind() != reflect.Map {
		return errors.Error("unmarshal: expected struct, got " + v.Kind().String())
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	for ik, iv := range m {
		ov := reflect.Indirect(reflect.New(v.Type().Elem()))
		err := unmarshalAny(iv.Hint(), iv, ov)
		if err != nil {
			return fmt.Errorf("unmarshal map: %w", err)
		}
		v.SetMapIndex(reflect.ValueOf(ik), ov)
	}
	return nil
}

func unmarshalAny(h hint.Hint, v value.Value, target reflect.Value) error {
	switch vv := v.(type) {
	case value.CID:
		if vv == "" {
			return nil
		}
		if target.Kind() != reflect.String {
			return errors.Error(
				"expected cid target, got " + target.Kind().String(),
			)
		}
		target.SetString(string(vv))
		return nil
	case value.String:
		// TODO is there any reason we would want to unmarshal an empty string?
		if vv == "" {
			return nil
		}
		// ie crypto.PublicKey
		if target.Kind() == reflect.Struct {
			if ivv, ok := target.Addr().Interface().(StringUnmashaller); ok {
				return ivv.UnmarshalString(string(vv))
			}
		}
		// ie *crypto.PublicKey
		if target.Kind() == reflect.Ptr {
			if _, ok := target.Interface().(StringUnmashaller); ok {
				ev := reflect.New(target.Type().Elem())
				itv := ev.Interface().(StringUnmashaller)
				target.Set(ev)
				return itv.UnmarshalString(string(vv))
			}
		}
		if target.Kind() != reflect.String {
			return errors.Error(
				"expected string target, got " + target.Kind().String(),
			)
		}
		target.SetString(string(vv))
		return nil
	case value.Bool:
		if target.Kind() != reflect.Bool {
			return errors.Error(
				"expected bool target, got " + target.Kind().String(),
			)
		}
		target.SetBool(bool(vv))
		return nil
	case value.Map:
		switch target.Kind() {
		case reflect.Ptr:
			target.Set(reflect.New(target.Type().Elem()))
		case reflect.Struct:
			target.Set(reflect.New(target.Type()).Elem())
		case reflect.Map:
			// if target.IsNil() {
			target.Set(reflect.New(target.Type()).Elem())
			// }
		default:
			return errors.Error(
				"expected map or struct target, got " + target.Kind().String(),
			)
		}
		return unmarshalMap(h, vv, target)
	case value.BoolArray,
		value.DataArray,
		value.FloatArray,
		value.IntArray,
		value.MapArray,
		// value.ObjectArray,
		value.StringArray,
		value.UintArray,
		value.CIDArray:
		switch target.Kind() {
		case reflect.Slice, reflect.Array:
		default:
			return errors.Error(
				"expected slice target, got " + target.Kind().String(),
			)
		}
		et := target.Type().Elem()
		var err error
		al := vv.(value.ArrayValue).Len()
		av := reflect.MakeSlice(target.Type(), 0, al)
		vv.(value.ArrayValue).Range(func(_ int, ov value.Value) bool {
			var ev reflect.Value
			if et.Kind() == reflect.Ptr {
				ev = reflect.Indirect(reflect.New(et))
			} else if et.Kind() == reflect.Struct {
				ev = reflect.Indirect(reflect.New(et).Elem())
			} else {
				ev = reflect.Indirect(reflect.New(et))
			}
			err = unmarshalAny(ov.Hint(), ov, ev)
			if err != nil {
				return false
			}
			av = reflect.Append(av, ev)
			return false
		})
		target.Set(av)
		return err
	case value.Float:
		switch target.Kind() {
		case reflect.Float32,
			reflect.Float64:
		default:
			return errors.Error(
				"expected float target, got " + target.Kind().String(),
			)
		}
		target.SetFloat(float64(vv))
		return nil
	case value.Int:
		switch target.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
		default:
			return errors.Error(
				"expected int target, got " + target.Kind().String(),
			)
		}
		target.SetInt(int64(vv))
		return nil
	case value.Uint:
		switch target.Kind() {
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
		default:
			return errors.Error(
				"expected uint target, got " + target.Kind().String(),
			)
		}
		target.SetUint(uint64(vv))
		return nil
	case value.Data:
		switch target.Kind() {
		case reflect.Ptr:
			if _, ok := target.Interface().(ByteUnmashaller); ok {
				ev := reflect.New(target.Type().Elem())
				itv := ev.Interface().(ByteUnmashaller)
				target.Set(ev)
				return itv.UnmarshalBytes([]byte(vv))
			}
		case reflect.Struct:
			if ivv, ok := target.Addr().Interface().(ByteUnmashaller); ok {
				return ivv.UnmarshalBytes([]byte(vv))
			}
		case reflect.Slice, reflect.Array:
		default:
			return errors.Error(
				"expected data target, got " + target.Kind().String(),
			)
		}
		target.SetBytes([]byte(vv))
		return nil
	}
	return nil
}
