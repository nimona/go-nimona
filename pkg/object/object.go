package object

import (
	"reflect"

	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/mapstructure"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
)

const (
	keyMetadata = "metadata:m"
	keyData     = "data:m"
	keyType     = "type:s"
)

const (
	ErrSourceNotSupported = errors.Error("encoding source not supported")
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
		Owner     crypto.PublicKey `nimona:"owner:s,omitempty"`
		Parents   []Hash           `nimona:"parents:as,omitempty"`
		Policy    Policy           `nimona:"policy:m,omitempty"`
		Stream    Hash             `nimona:"stream:s,omitempty"`
		Signature Signature        `nimona:"_signature:m,omitempty"`
	}
	// Policy for object metadata
	Policy struct {
		Subjects  []string `nimona:"subjects:as,omitempty"`
		Resources []string `nimona:"resources:as,omitempty"`
		Actions   []string `nimona:"actions:as,omitempty"`
		Effect    string   `nimona:"effect:s,omitempty"`
	}
)

func FromMap(m map[string]interface{}) *Object {
	o, err := mapToObject(m)
	if err != nil {
		panic(err)
	}
	return o
}

func (o Object) ToMap() map[string]interface{} {
	m, err := objectToMap(&o)
	if err != nil {
		panic(err)
	}
	return m
}

func (o *Object) Hash() Hash {
	h, err := NewHash(o)
	if err != nil {
		panic(err)
	}
	return h
}

func (h Hash) String() string {
	return string(h)
}

func Encode(v interface{}) (*Object, error) {
	m := map[string]interface{}{}
	switch vi := v.(type) {
	case *Object:
		return vi, nil
	case Typed:
		d := map[string]interface{}{}
		if _, err := decode(v, &d, encodeHookfunc()); err != nil {
			return nil, err
		}
		m = map[string]interface{}{
			keyType:     vi.Type(),
			keyData:     d,
			keyMetadata: d[keyMetadata],
		}
		delete(d, keyMetadata)
	case map[string]interface{}:
		m = vi
	case map[interface{}]interface{}:
		if _, err := decode(vi, &m, nilHookfunc()); err != nil {
			return nil, err
		}
	default:
		return nil, ErrSourceNotSupported
	}
	return mapToObject(m)
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

func nilHookfunc() mapstructure.DecodeHookFuncValueContext {
	return func(
		f reflect.Value,
		t reflect.Value,
		ctx *mapstructure.DecodeContext,
	) (interface{}, error) {
		return f.Interface(), nil
	}
}

func encodeHookfunc() mapstructure.DecodeHookFuncValueContext {
	topLevelTyped := true
	return func(
		f reflect.Value,
		t reflect.Value,
		ctx *mapstructure.DecodeContext,
	) (interface{}, error) {
		// (encode) crypto.PrivateKey
		if v, ok := f.Interface().(crypto.PrivateKey); ok {
			return string(v), nil
		}
		// (encode) crypto.PublicKey
		if v, ok := f.Interface().(crypto.PublicKey); ok {
			return string(v), nil
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
		// (encode) map[string]interface{} with type/data/metadat to *Object
		// simpler things
		// if !ctx.IsKey {
		// 	r, err := normalizeFromKey(ctx.Name, f.Interface())
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	return r, nil
		// }
		return f.Interface(), nil
	}
}

func decodeHookfunc() mapstructure.DecodeHookFuncValueContext {
	topLevelTyped := true
	return func(
		f reflect.Value,
		t reflect.Value,
		ctx *mapstructure.DecodeContext,
	) (interface{}, error) {
		// (decode) crypto.PrivateKey
		if v, ok := f.Interface().(crypto.PrivateKey); ok {
			return string(v), nil
		}
		// (decode) crypto.PublicKey
		if v, ok := f.Interface().(crypto.PublicKey); ok {
			return string(v), nil
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
				switch fvv := fv.(type) {
				case *Object:
					if _, err := decode(fvv, &tv, decodeHookfunc()); err != nil {
						return nil, err
					}
					if _, err := decode(fvv.Data, &tv, decodeHookfunc()); err != nil {
						return nil, err
					}
					sliceValuePtr.Set(reflect.Append(sliceValuePtr, reflect.ValueOf(tv)))
				case map[string]interface{}:
					if _, err := decode(fvv, &tv, decodeHookfunc()); err != nil {
						return nil, err
					}
					md, ok := fvv[keyData]
					if ok {
						if _, err := decode(md, &tv, decodeHookfunc()); err != nil {
							return nil, err
						}
					}
					sliceValuePtr.Set(reflect.Append(sliceValuePtr, reflect.ValueOf(tv)))
				}
			}
			return slicePtr.Interface(), nil
		}
		// (decode) slice of struct to []interface
		if f.Kind() == reflect.Slice {
			var i interface{} = struct{}{}
			if t.Type() != reflect.TypeOf(&i).Elem() {
				return f.Interface(), nil
			}
			m := make([]interface{}, 0)
			t.Set(reflect.ValueOf(m))
			return f.Interface(), nil
		}
		// (decode) struct to map[string]interface{}
		// mapstructure.RecursiveStructToMapHookFunc
		var i interface{} = struct{}{}
		if f.Kind() == reflect.Struct && t.Type() == reflect.TypeOf(&i).Elem() {
			m := make(map[string]interface{})
			t.Set(reflect.ValueOf(m))
			return f.Interface(), nil
		}
		return f.Interface(), nil
	}
}

func mapHookfunc() mapstructure.DecodeHookFuncValueContext {
	return func(
		f reflect.Value,
		t reflect.Value,
		ctx *mapstructure.DecodeContext,
	) (interface{}, error) {
		// (decode) struct to map[string]interface{}
		// mapstructure.RecursiveStructToMapHookFunc
		if f.Kind() == reflect.Slice ||
			(f.Kind() == reflect.Ptr && f.Elem().Kind() == reflect.Slice) {
			var i interface{} = struct{}{}
			if t.Type() != reflect.TypeOf(&i).Elem() {
				return f.Interface(), nil
			}
			m := make([]interface{}, 0)
			t.Set(reflect.ValueOf(m))
		}
		// (decode) crypto.PrivateKey
		if v, ok := f.Interface().(crypto.PrivateKey); ok {
			return string(v), nil
		}
		// (decode) crypto.PublicKey
		if v, ok := f.Interface().(crypto.PublicKey); ok {
			return string(v), nil
		}
		// (decode) *Object
		if v, ok := f.Interface().(*Object); ok {
			return v.ToMap(), nil
		}
		// (decode) slice of struct to []interface
		if f.Kind() == reflect.Slice {
			var i interface{} = struct{}{}
			if t.Type() != reflect.TypeOf(&i).Elem() {
				return f.Interface(), nil
			}
			m := make([]interface{}, 0)
			t.Set(reflect.ValueOf(m))
			return f.Interface(), nil
		}
		return f.Interface(), nil
	}
}

func decode(
	from interface{},
	to interface{},
	hook mapstructure.DecodeHookFuncValueContext,
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

func objectToMap(o *Object) (map[string]interface{}, error) {
	d := map[string]interface{}{}
	m := map[string]interface{}{}
	if _, err := decode(o.Data, &d, mapHookfunc()); err != nil {
		return nil, err
	}
	if _, err := decode(o.Metadata, &m, mapHookfunc()); err != nil {
		return nil, err
	}
	r := map[string]interface{}{
		keyType:     o.Type,
		keyData:     d,
		keyMetadata: m,
	}
	return r, nil
}

func mapToObject(m map[string]interface{}) (*Object, error) {
	t := ""
	if ti, ok := m[keyType]; ok {
		t = ti.(string)
	}
	o := &Object{
		Type:     t,
		Metadata: Metadata{},
		Data:     map[string]interface{}{},
	}
	if mm, ok := m[keyData]; ok {
		d, err := normalizeFromKey(":m", mm)
		if err != nil {
			return nil, err
		}
		if _, err := decode(d, &o.Data, encodeHookfunc()); err != nil {
			return nil, err
		}
	}
	if mm, ok := m[keyMetadata]; ok {
		d, err := normalizeFromKey(":m", mm)
		if err != nil {
			return nil, err
		}
		if _, err := decode(d, &o.Metadata, encodeHookfunc()); err != nil {
			return nil, err
		}
	}
	return o, nil
}

func Copy(s *Object) *Object {
	r, err := copystructure.Copy(s)
	if err != nil {
		panic(err)
	}
	return r.(*Object)
}
