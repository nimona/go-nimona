package object

import (
	"fmt"
	"reflect"

	"nimona.io/pkg/errors"
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
	return unmarshalMap(MapHint, o.Data, v)
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

func unmarshalMap(h Hint, m Map, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return unmarshalMapToStruct(h, m, v)
	case reflect.Map:
		return errors.Error("maps are not currently supported")
	}

	return errors.Error("expected map or struct, got " + v.Kind().String())
}

func unmarshalMapToStruct(h Hint, m Map, v reflect.Value) error {
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
		in, ih, err := splitHint([]byte(ig))
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

func unmarshalAny(h Hint, v Value, target reflect.Value) error {
	if !target.CanSet() {
		return fmt.Errorf("cannot set value")
	}
	switch vv := v.(type) {
	case CID:
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
	case String:
		// TODO is there any reason we would want to unmarshal an empty string?
		if vv == "" {
			return nil
		}
		// ie crypto.PublicKey
		if ivv, ok := target.Addr().Interface().(StringUnmashaller); ok {
			return ivv.UnmarshalString(string(vv))
		}
		// ie *crypto.PublicKey
		if _, ok := target.Interface().(StringUnmashaller); ok {
			ev := reflect.New(target.Type().Elem())
			itv := ev.Interface().(StringUnmashaller)
			target.Set(ev)
			return itv.UnmarshalString(string(vv))
		}
		if target.Kind() != reflect.String {
			return errors.Error(
				"expected string target, got " + target.Kind().String(),
			)
		}
		target.SetString(string(vv))
		return nil
	case Bool:
		if target.Kind() != reflect.Bool {
			return errors.Error(
				"expected bool target, got " + target.Kind().String(),
			)
		}
		target.SetBool(bool(vv))
		return nil
	case Map:
		switch target.Kind() {
		case reflect.Struct, reflect.Map:
		default:
			return errors.Error(
				"expected map or struct target, got " + target.Kind().String(),
			)
		}
		return unmarshalMap(h, vv, target)
	case *Object:
		var ev reflect.Value
		if target.Kind() == reflect.Ptr {
			ev = reflect.New(target.Type().Elem())
		} else {
			ev = reflect.New(target.Type())
		}
		// if the target is an object simply set it
		if _, ok := ev.Interface().(*Object); ok {
			target.Set(reflect.ValueOf(vv))
			return nil
		}
		// else we should try to unmarshal onto it
		err := Unmarshal(vv, ev.Interface())
		if err != nil {
			return err
		}
		if target.Kind() == reflect.Ptr {
			target.Set(ev)
		} else {
			target.Set(ev.Elem())
		}
		return nil
	case BoolArray,
		DataArray,
		FloatArray,
		IntArray,
		MapArray,
		ObjectArray,
		StringArray,
		UintArray,
		CIDArray:
		switch target.Kind() {
		case reflect.Slice, reflect.Array:
		default:
			return errors.Error(
				"expected slice target, got " + target.Kind().String(),
			)
		}
		et := target.Type().Elem()
		var err error
		vv.(ArrayValue).Range(func(_ int, ov Value) bool {
			ev := reflect.Indirect(reflect.New(et))
			err = unmarshalAny(ov.Hint(), ov, ev)
			if err != nil {
				return true
			}
			target.Set(reflect.Append(target, ev))
			return false
		})
		return err
	case Float:
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
	case Int:
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
	case Uint:
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
	case Data:
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
	// switch h {
	// case ObjectHint:
	// 	ev := reflect.New(target.Type().Elem())
	// 	fmt.Println("!!!", h, ev.Elem().Kind().String())
	// 	Unmarshal(v, out interface{})
	// }
	return nil
}
