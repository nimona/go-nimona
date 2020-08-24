package objectv3

import (
	"crypto"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

const (
	keyMetadata = "metadata:m"
)

type (
	Typed interface {
		Type() string
	}
	Hash   string
	Object struct {
		Type     string                 `nimona:"type:s,omitempty"`
		Metadata Metadata               `nimona:"metadata:m,omitempty"`
		Data     map[string]interface{} `nimona:"data:m,omitempty"`
	}
	// Metadata for object
	Metadata struct {
		Owner   crypto.PublicKey `nimona:"owner:s,omitempty"`
		Parents []Hash           `nimona:"parents:as,omitempty"`
		Policy  Policy           `nimona:"policy:m,omitempty"`
		Stream  Hash             `nimona:"stream:s,omitempty"`
	}
	// Policy for object metadata
	Policy struct {
		Subjects  []string `nimona:"subjects:as,omitempty"`
		Resources []string `nimona:"resources:as,omitempty"`
		Actions   []string `nimona:"actions:as,omitempty"`
		Effect    string   `nimona:"effect:s,omitempty"`
	}
)

func Encode(v Typed) (*Object, error) {
	m := map[string]interface{}{}
	if _, err := decode(v, &m, encodeHookfunc()); err != nil {
		return nil, err
	}
	o := &Object{
		Type:     v.Type(),
		Metadata: Metadata{},
		Data:     m,
	}
	if mm, ok := m[keyMetadata]; ok {
		if _, err := decode(mm, &o.Metadata, encodeHookfunc()); err != nil {
			return nil, err
		}
		delete(m, keyMetadata)
	}
	return o, nil
}

func Decode(o *Object, v Typed) error {
	// TODO check type equality
	if _, err := decode(o, v, decodeHookfunc()); err != nil {
		return err
	}
	if _, err := decode(o.Data, v, decodeHookfunc()); err != nil {
		return err
	}
	return nil
}

var (
	typeOfTyped  = reflect.TypeOf((*Typed)(nil)).Elem()
	typeOfObject = reflect.TypeOf((*Object)(nil)).Elem()
)

func encodeHookfunc() mapstructure.DecodeHookFuncValue {
	topLevelTyped := true
	return func(
		f reflect.Value,
		t reflect.Value,
	) (interface{}, error) {
		if t.Type() == reflect.TypeOf(reflect.Value{}) {
			return f.Interface(), nil
		}
		if f.Type() == reflect.TypeOf(reflect.Value{}) {
			return f.Interface(), nil
		}
		// (encode) Typed to *Object
		if _, ok := f.Interface().(Typed); ok {
			if topLevelTyped {
				topLevelTyped = false
				return f.Interface(), nil
			}
			o, err := Encode(f.Interface().(Typed))
			if err != nil {
				return nil, err
			}
			return o, nil
		}
		// (encode) []Typed to []*Object
		if f.Kind() == reflect.Slice &&
			f.Type().Elem().Implements(typeOfTyped) {
			os := make([]*Object, f.Len())
			for i := 0; i < f.Len(); i++ {
				o, err := Encode(f.Index(i).Interface().(Typed))
				if err != nil {
					return nil, err
				}
				os[i] = o
			}
			return os, nil
		}
		return f.Interface(), nil
	}
}

func decodeHookfunc() mapstructure.DecodeHookFuncValue {
	topLevelTyped := true
	return func(
		f reflect.Value,
		t reflect.Value,
	) (interface{}, error) {
		if t.Type() == reflect.TypeOf(reflect.Value{}) {
			return f.Interface(), nil
		}
		if f.Type() == reflect.TypeOf(reflect.Value{}) {
			return f.Interface(), nil
		}
		// (decode) *Object to Typed
		if _, ok := t.Interface().(Typed); ok &&
			f.Type().Elem() == typeOfObject {
			if topLevelTyped {
				topLevelTyped = false
				return f.Interface(), nil
			}
			tt := t.Type()
			if tt.Kind() == reflect.Ptr {
				tt = t.Type().Elem()
			}
			tv := reflect.New(tt).Interface()
			err := Decode(f.Interface().(*Object), tv.(Typed))
			if err != nil {
				return nil, err
			}
			return tv, nil
		}
		// (decode) []*Object to []Typed
		if t.Kind() == reflect.Slice &&
			t.Type().Elem().Implements(typeOfTyped) {
			reflection := reflect.MakeSlice(reflect.SliceOf(t.Type().Elem()), 0, 0)
			reflectionValue := reflect.New(reflection.Type())
			reflectionValue.Elem().Set(reflection)
			slicePtr := reflect.ValueOf(reflectionValue.Interface())
			sliceValuePtr := slicePtr.Elem()
			for i := 0; i < f.Len(); i++ {
				fv := f.Index(i).Interface()
				tv := reflect.Zero(t.Type().Elem()).Interface()
				if _, err := decode(fv.(*Object), &tv, decodeHookfunc()); err != nil {
					return nil, err
				}
				if _, err := decode(fv.(*Object).Data, &tv, decodeHookfunc()); err != nil {
					return nil, err
				}
				sliceValuePtr.Set(reflect.Append(sliceValuePtr, reflect.ValueOf(tv)))
			}
			return slicePtr.Interface(), nil
		}
		return f.Interface(), nil
	}
}

func decode(
	from interface{},
	to interface{},
	hook mapstructure.DecodeHookFuncValue,
) (*mapstructure.Metadata, error) {
	md := &mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		DecodeHook:       hook,
		ErrorUnused:      false,
		ZeroFields:       false,
		WeaklyTypedInput: true,
		Squash:           false,
		Metadata:         md,
		Result:           to,
		TagName:          "nimona",
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return md, err
	}
	if err := decoder.Decode(from); err != nil {
		return md, err
	}
	return md, nil
}
