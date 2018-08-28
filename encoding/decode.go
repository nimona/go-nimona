package encoding

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func Decode(b *Block, r interface{}) {
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
	block := s.block
	payload := s.block.Payload.(map[string]interface{})

	cfg := &mapstructure.DecoderConfig{
		TagName:          DefaultTagName,
		ZeroFields:       false,
		WeaklyTypedInput: true,
		ErrorUnused:      false,
		Result:           s.raw,
		DecodeHook: func(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
			if from.Kind() == reflect.Slice && to == reflect.TypeOf(&Signature{}) {
				s := &Signature{}
				err := Unmarshal(v.([]byte), s)
				return s, err
			}
			return v, nil
		},
	}
	dec, _ := mapstructure.NewDecoder(cfg)
	if err := dec.Decode(payload); err != nil {
		panic(err)
	}

	// now deal with magic fields on the first level of the dest
	// TODO probably this should be feasible by just using mapstructure
	fields := s.structFields()
	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)

		tagName, tagOpts := parseTag(field.Tag.Get(s.TagName))
		if tagName != "" {
			name = tagName
		}

		if tagOpts.Has("signature") {
			sig := &Signature{}
			err := Unmarshal(s.block.Signature, sig)
			if err != nil {
				panic(err)
			}
			vval := reflect.ValueOf(sig)
			val.Set(vval)
			continue
		}

		if tagOpts.Has("parent") {
			if block.Metadata == nil {
				block.Metadata = &Metadata{}
			}
			block.Metadata.Parent = val.Interface().(string)
			continue
		}

		if tagOpts.Has("type") {
			vval := reflect.ValueOf(s.block.Type)
			val.Set(vval)
			continue
		}
	}
}
