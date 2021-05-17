package object

import (
	"fmt"
	"reflect"

	"nimona.io/pkg/errors"
)

func Unmarshal(o *Object, out interface{}) error {
	v := reflect.ValueOf(out)
	err := unmarshalSpecials(o, v)
	if err != nil {
		return err
	}
	return unmarshalMap(o.Map(), v)
}

func unmarshalSpecials(o *Object, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.Error("expected struct, got " + v.Kind().String())
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
			return fmt.Errorf("attribute %s, %w", it.Name, err)
		}
		switch ig {
		case "@type:s":
			iv.Set(reflect.ValueOf(o.Type))
		case "@metadata:m":
			iv.Set(reflect.ValueOf(o.Metadata))
		}
	}
	return nil
}

func unmarshalMap(m Map, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return unmarshalMapToStruct(m, v)
	case reflect.Map:
		return unmarshalMapToMap(m, v)
	}

	return errors.Error("expected map or struct, got " + v.Kind().String())
}

func unmarshalMapToStruct(m Map, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors.Error("expected struct, got " + v.Kind().String())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		iv := v.Field(i)
		it := t.Field(i)
		if it.Anonymous {
			return unmarshalMapToStruct(m, iv)
		}
		ig, err := getStructTagName(it)
		if err != nil {
			return fmt.Errorf("attribute %s, %w", it.Name, err)
		}
		switch ig {
		case "@type:s",
			"@metadata:m":
			// TODO special
			continue
		}
		in, _, err := splitHint([]byte(ig))
		if err != nil {
			return err
		}
		val, ok := m[in]
		if !ok {
			continue
		}
		if err := unmarshalAny(val, iv); err != nil {
			return err
		}
	}
	return nil
}

// TODO(geoah): Might need to implement at some point, not sure what the
// usecase is though.
// nolint: gocritic
func unmarshalMapToMap(m Map, v reflect.Value) error {
	return errors.Error("maps are not currently supported")
	// 	if v.Kind() == reflect.Ptr {
	// 		v = v.Elem()
	// 	}
	// 	if v.Kind() != reflect.Map {
	// 		return errors.Error("expected struct, got " + v.Kind().String())
	// 	}

	// 	if v.IsNil() {
	// 		v.Set(reflect.MakeMap(v.Type()))
	// 	}

	// 	et := v.Type().Elem()
	// 	for in, ov := range m {
	// 		ig := in + ":" + string(ov.Hint())
	// 		ev := reflect.Indirect(reflect.New(et))
	// 		if err := unmarshalAny(ov, ev); err != nil {
	// 			return err
	// 		}
	// 		v.SetMapIndex(reflect.ValueOf(ig), ev)
	// 	}
	// 	return nil
}

func unmarshalAny(v Value, target reflect.Value) error {
	if !target.CanSet() {
		return fmt.Errorf("cannot set value")
	}
	switch vv := v.(type) {
	case String:
		if target.Kind() != reflect.String {
			return errors.Error(
				"expected string target, got " + target.Kind().String(),
			)
		}
		target.SetString(string(vv))
	case Bool:
		if target.Kind() != reflect.Bool {
			return errors.Error(
				"expected bool target, got " + target.Kind().String(),
			)
		}
		target.SetBool(bool(vv))
	case Map:
		switch target.Kind() {
		case reflect.Struct, reflect.Map:
		default:
			return errors.Error(
				"expected map or struct target, got " + target.Kind().String(),
			)
		}
		return unmarshalMap(vv, target)
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
			err = unmarshalAny(ov, ev)
			if err != nil {
				return false
			}
			target.Set(reflect.Append(target, ev))
			return true
		})
		if err != nil {
			return err
		}
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
	}
	return nil
}
