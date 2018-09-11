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
	// HACK we currently always base58 encode nested blocks
	opts = append(opts, EncodeNestedBase58())
	opts = append(opts, EncodeNested())

	o := ParsePackOptions(opts...)
	if o.Sign && o.Key != nil {
		opts = append(opts, SignWith(nil))
	}
	t := v.GetType()
	if t == "" {
		return nil, errors.New("empty type")
	}
	p, err := toMap(v, tagName, opts...)
	if err != nil {
		return nil, err
	}
	b := &Block{
		Type:    t,
		Payload: p,
	}
	if o.Sign && o.Key != nil {
		sig, err := signPacked(b, o.Key)
		if err != nil {
			return nil, err
		}
		b.Signature = sig
	} else if s := v.GetSignature(); s != nil {
		ss, err := PackEncodeBase58(s)
		if err != nil {
			return nil, err
		}
		b.Signature = ss
	}
	return b, nil
}

// TODO support nested structs etc
// TODO support for ,omitempty
func toMap(in interface{}, tag string, opts ...PackOption) (map[string]interface{}, error) {
	o := ParsePackOptions(opts...)

	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() {
		return nil, nil
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("toMap only accepts structs; got %s", v.Kind())
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv := fi.Tag.Get(tag); tagv != "" {
			tagName, tagOpts := parseTag(tagv)
			if tagName == "-" {
				continue
			}
			fv := v.Field(i)
			if !fv.IsValid() || tagOpts.Has("omitempty") && isEmptyValue(fv) {
				continue
			}
			// if implements Typed, pack it
			if reflect.TypeOf(fv.Interface()).Implements(typedType) {
				if v.Field(i).IsNil() {
					continue
				}
				var nv interface{}
				var err error
				iv := v.Field(i).Interface()
				if o.EncodeNestedBase58 {
					nv, err = PackEncodeBase58(iv.(Typed), opts...)
				} else if o.EncodeNested {
					nv, err = PackEncode(iv.(Typed), opts...)
				} else {
					nv, err = Pack(iv.(Typed), opts...)
				}
				if err != nil {
					return nil, err
				}
				out[tagName] = nv
				continue
			}
			// else set key of map to value in struct field
			out[tagName] = v.Field(i).Interface()
		}
	}
	return out, nil
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
