package blocks

import (
	"reflect"
	"unicode"

	"github.com/mitchellh/mapstructure"
)

func Decode(o *Block) {
	t := GetType(o.Type)
	pt := reflect.PtrTo(t)
	v := reflect.New(pt).Elem().Interface()
	rv := reflect.ValueOf(&v).Elem()
	rvt := rv.Elem().Type().Elem()
	rv.Set(reflect.New(rvt))

	DecodeInto(o, v)
	o.Payload = v
}

func DecodeInto(b *Block, r interface{}) {
	// TODO catch panics
	s := &Struct{
		raw:     r,
		value:   strctVal(r),
		block:   b,
		TagName: DefaultTagName,
	}
	s.Decode()
}

// Decode block and its payload into user-specified struct
func (s *Struct) Decode() {
	payload := s.block.Payload.(map[string]interface{})

	cfg := &mapstructure.DecoderConfig{
		TagName:          DefaultTagName,
		ZeroFields:       false,
		WeaklyTypedInput: true,
		ErrorUnused:      false,
		Result:           s.raw,
		DecodeHook: func(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
			if v == nil {
				return nil, nil
			}
			// TODO don't think the signature and key cases are needed any more
			if from.Kind() == reflect.Slice && to == reflect.TypeOf(&Signature{}) {
				s := &Signature{}
				err := UnmarshalInto(v.([]byte), s)
				return s, err
			}
			if from.Kind() == reflect.Slice && to == reflect.TypeOf(&Key{}) {
				s := &Key{}
				err := UnmarshalInto(v.([]byte), s)
				return s, err
			}
			if from.Kind() == reflect.Slice && to.Implements(marshalerType) {
				s := TypeToInterface(to)
				err := UnmarshalInto(v.([]byte), s)
				return s, err
			}
			return v, nil
		},
	}

	dec, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		panic(err)
	}

	if err := dec.Decode(payload); err != nil {
		panic(err)
	}

	// now deal with magic fields on the first level of the dest
	// TODO probably this should be feasible by just using mapstructure
	fields := s.structFields()
	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)

		if unicode.IsLower(rune(name[0])) {
			continue
		}

		tagName, tagOpts := parseTag(field.Tag.Get(s.TagName))
		if tagName != "" {
			name = tagName
		}

		if tagOpts.Has("header") {
			v, ok := s.block.Headers[tagName]
			if !ok {
				continue
			}
			vval := reflect.ValueOf(v)
			val.Set(vval)
			continue
		}

		if tagOpts.Has("signature") {
			sig := &Signature{}
			err := UnmarshalInto(s.block.Signature, sig)
			if err != nil {
				panic(err)
			}
			vval := reflect.ValueOf(sig)
			val.Set(vval)
			continue
		}

		if tagOpts.Has("parent") {
			if s.block.Metadata == nil {
				continue
			}
			if s.block.Metadata.Parent == "" {
				continue
			}
			vval := reflect.ValueOf(s.block.Metadata.Parent)
			val.Set(vval)
			continue
		}

		if tagOpts.Has("type") {
			vval := reflect.ValueOf(s.block.Type)
			val.Set(vval)
			continue
		}
	}
}
