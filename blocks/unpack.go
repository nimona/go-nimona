package blocks

import (
	"errors"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"nimona.io/go/base58"
	"nimona.io/go/codec"
	"nimona.io/go/crypto"
)

func UnpackDecodeBase58(v string, opts ...UnpackOption) (Typed, error) {
	opts = append(opts, DecodeNestedBase58())
	b, err := base58.Decode(v)
	if err != nil {
		return nil, err
	}
	return UnpackDecode(b, opts...)
}

func DecodeBase58(v string, opts ...UnpackOption) (*Block, error) {
	opts = append(opts, DecodeNestedBase58())
	b, err := base58.Decode(v)
	if err != nil {
		return nil, err
	}
	return Decode(b, opts...)
}

func UnpackDecode(b []byte, opts ...UnpackOption) (Typed, error) {
	opts = append(opts, DecodeNested())
	p := &Block{}
	if err := codec.Unmarshal(b, p); err != nil {
		return nil, err
	}
	return Unpack(p)
}

func Decode(b []byte, opts ...UnpackOption) (*Block, error) {
	opts = append(opts, DecodeNested())
	p := &Block{}
	if err := codec.Unmarshal(b, p); err != nil {
		return nil, err
	}
	return p, nil
}

func Unpack(p *Block, opts ...UnpackOption) (Typed, error) {
	ts := p.Type
	if ts == "" {
		return nil, errors.New("missing type")
	}
	t := GetType(ts)
	if t == nil {
		panic("missing type")
	}
	v := TypeToInterface(t).(Typed)
	if err := UnpackInto(p, v); err != nil {
		return nil, err
	}
	return v.(Typed), nil
}

func blockMapToBlock(m map[string]interface{}) (*Block, error) {
	b := &Block{}
	md := &mapstructure.DecoderConfig{
		TagName:          defaultTag,
		ZeroFields:       false,
		WeaklyTypedInput: true,
		ErrorUnused:      false,
		Result:           b,
	}
	d, err := mapstructure.NewDecoder(md)
	if err != nil {
		return nil, err
	}
	if err := d.Decode(m); err != nil {
		return nil, err
	}
	return b, nil
}

// UnpackInto unpacks a Block into a given Typed
func UnpackInto(p *Block, v Typed, opts ...UnpackOption) error {
	o := ParseUnpackOptions(opts...)
	if o.Verify && p.Signature != nil {
		// TODO verify signature
		d, err := getDigest(p)
		if err != nil {
			return err
		}
		bs, err := blockMapToBlock(p.Signature)
		if err != nil {
			return err
		}
		s, err := Unpack(bs)
		if err != nil {
			return err
		}
		if err := crypto.Verify(s.(*crypto.Signature), d); err != nil {
			return err
		}
	}
	md := &mapstructure.DecoderConfig{
		TagName:          defaultTag,
		ZeroFields:       false,
		WeaklyTypedInput: true,
		ErrorUnused:      false,
		Result:           v,
		DecodeHook: func(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
			// if from.Kind() == reflect.String && to.Implements(typedType) {
			// 	iv, err := UnpackDecodeBase58(v.(string), opts...)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// 	return iv, nil
			// }
			// if from.Kind() == reflect.Slice && to.Implements(typedType) {
			// 	iv, err := UnpackDecode(v.([]byte), opts...)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// 	return iv, nil
			// }
			if from.Kind() == reflect.Map && to.Implements(typedType) {
				b, err := blockMapToBlock(v.(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				t, err := Unpack(b)
				if err != nil {
					return nil, err
				}
				return t, nil
			}
			return v, nil
		},
	}
	d, err := mapstructure.NewDecoder(md)
	if err != nil {
		return err
	}
	if err := d.Decode(p.Payload); err != nil {
		return err
	}
	if p.Signature != nil {
		// TODO deduplicate code
		bs, err := blockMapToBlock(p.Signature)
		if err != nil {
			return err
		}
		s, err := Unpack(bs)
		if err != nil {
			return err
		}
		v.SetSignature(s.(*crypto.Signature))
	}
	return nil
}

// used for encoding as all registered types should be strcts
func TypeToPtrInterface(t reflect.Type) interface{} {
	pt := reflect.PtrTo(t)
	v := reflect.New(pt).Elem().Interface()
	rv := reflect.ValueOf(&v).Elem()
	rvt := rv.Elem().Type().Elem()
	rv.Set(reflect.New(rvt))
	return v
}

// used for decoding as all payloads are ptrs already
func TypeToInterface(t reflect.Type) interface{} {
	// pt := reflect.PtrTo(t)
	v := reflect.New(t).Elem().Interface()
	rv := reflect.ValueOf(&v).Elem()
	rvt := rv.Elem().Type().Elem()
	rv.Set(reflect.New(rvt))
	return v
}
