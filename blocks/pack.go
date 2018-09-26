package blocks

import (
	"errors"
	"fmt"
	"reflect"

	"nimona.io/go/base58"
	"nimona.io/go/codec"
)

func PackEncodeBase58(v Typed, opts ...PackOption) (string, error) {
	opts = append(opts, EncodeNestedBase58())
	b, err := PackEncode(v, opts...)
	if err != nil {
		return "", err
	}
	return base58.Encode(b), nil
}

func EncodeBase58(p *Block) (string, error) {
	b, err := Encode(p)
	if err != nil {
		return "", err
	}
	return base58.Encode(b), nil
}

func PackEncode(v Typed, opts ...PackOption) ([]byte, error) {
	opts = append(opts, EncodeNested())
	p, err := Pack(v, opts...)
	if err != nil {
		return nil, err
	}
	return codec.Marshal(p)
}

func Encode(p *Block) ([]byte, error) {
	return codec.Marshal(p)
}

// Pack gets something Typed and converts it into a Block
func Pack(v Typed, opts ...PackOption) (*Block, error) {
	o := ParsePackOptions(opts...)
	if o.Sign && o.Key != nil {
		opts = append(opts, SignWith(nil))
	}
	m, err := MapTyped(v)
	if err != nil {
		return nil, err
	}
	b := &Block{
		Type:    m["type"].(string),
		Payload: m["payload"].(map[string]interface{}),
	}
	if _, ok := m["annotations"]; ok {
		b.Annotations = m["annotations"].(map[string]interface{})
	}
	if _, ok := m["signature"]; ok {
		b.Signature = m["signature"].(map[string]interface{})
	}
	if o.Sign && o.Key != nil {
		s, err := SignPacked(b, o.Key)
		if err != nil {
			return nil, err
		}
		ps, err := MapTyped(s)
		if err != nil {
			return nil, err
		}
		b.Signature = ps
	}
	return b, nil
}

// MapTyped gets a Typed and converts it into a Map
func MapTyped(v Typed) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	t := v.GetType()
	if t == "" {
		return nil, errors.New("empty type")
	}
	m["type"] = t
	p, err := MapStruct(v)
	if err != nil {
		return nil, err
	}
	m["payload"] = p
	s := v.GetSignature()
	if s != nil {
		ps, err := MapTyped(s)
		if err != nil {
			return nil, err
		}
		m["signature"] = ps
	}
	m["annotations"] = v.GetAnnotations()
	return m, nil
}

func MapStruct(in interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("toMap only accepts structs; got %s", v.Kind())
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fi := typ.Field(i)
		if tagv := fi.Tag.Get(defaultTag); tagv != "" {
			tagName, tagOpts := parseTag(tagv)
			if tagName == "-" {
				continue
			}
			fv := v.Field(i)
			if !fv.IsValid() || tagOpts.Has("omitempty") && isEmptyValue(fv) {
				continue
			}
			if tagOpts.Has("payload") {
				if fv.Kind() == reflect.Struct {
					m, err := MapStruct(fv.Interface())
					if err != nil {
						panic(err)
					}
					for k, v := range m {
						out[k] = v
					}
					continue
				} else if fv.Kind() == reflect.Map {
					m := fv.Interface().(map[string]interface{})
					for k, v := range m {
						out[k] = v
					}
					continue
				}
			}
			mv, err := Map(fv.Interface())
			if err != nil {
				panic(err)
			}
			if mv == nil {
				continue
			}
			out[tagName] = mv
		}
	}
	if len(out) == 0 {
		return nil, nil
	}

	return out, nil
}

func Map(in interface{}) (interface{}, error) {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() {
		return nil, nil
	}

	if isEmptyValue(v) {
		return nil, nil
	}

	// if isEmptyValue(v) {
	// 	return nil, nil
	// }

	if reflect.TypeOf(in).Implements(typedType) {
		return MapTyped(in.(Typed))
	}

	if v.Kind() == reflect.Struct {
		return MapStruct(v.Interface())
	}

	if v.Kind() == reflect.Slice {
		sType := reflect.TypeOf(in).Elem().Kind()
		if sType == reflect.Ptr {
			sType = reflect.TypeOf(in).Elem().Elem().Kind()
		}
		if sType != reflect.Struct {
			return in, nil
		}
		sv := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			pv, err := Map(v.Index(i).Interface())
			if err != nil {
				panic(err)
			}
			sv = append(sv, pv)
		}
		return sv, nil
	}

	return v.Interface(), nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
