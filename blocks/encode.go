package blocks

import (
	"reflect"
)

func Encode(s interface{}) *Block {
	// TODO catch panics
	return New(s).Block()
}

type Struct struct {
	raw     interface{}
	block   *Block
	payload map[string]interface{}
	value   reflect.Value
	TagName string
}

// New returns a new *Struct with the struct s. It panics if the s's kind is
// not struct.
func New(s interface{}) *Struct {
	return &Struct{
		raw:   s,
		value: strctVal(s),
		block: &Block{
			Payload: map[string]interface{}{},
		},
		TagName: DefaultTagName,
	}
}

func strctVal(s interface{}) reflect.Value {
	v := reflect.ValueOf(s)

	// if pointer get the underlying element≤
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("not struct")
	}

	return v
}

func (s *Struct) Map() map[string]interface{} {
	s.Fill()
	return s.block.Payload.(map[string]interface{})
}

func (s *Struct) Block() *Block {
	s.Fill()
	return s.block
}

func (s *Struct) Fill() {
	// TODO get type from p and set to block

	ct := ""
	// TODO is the best place to set the type?
	rt := reflect.TypeOf(s.value.Interface())
	if cp := GetFromType(rt); cp != "" {
		ct = cp
	}

	block := s.block
	out := s.block.Payload.(map[string]interface{})

	block.Type = ct

	fields := s.structFields()

	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)
		// isSubStruct := false
		var finalVal interface{}

		tagName, tagOpts := parseTag(field.Tag.Get(s.TagName))
		if tagName != "" {
			name = tagName
		}

		if tagOpts.Has("signature") {
			if val.IsNil() {
				continue
			}
			sig, err := Marshal(val.Interface().(*Signature))
			if err != nil {
				panic(err)
			}
			block.Signature = Base58Encode(sig)
			continue
		}

		// if tagOpts.Has("header") {
		// 	if val.Interface().(string) == "" {
		// 		continue
		// 	}
		// 	if block.Headers == nil {
		// 		block.Headers = map[string]string{}
		// 	}
		// 	block.Headers[tagName] = val.Interface().(string)
		// 	continue
		// }

		if tagOpts.Has("parent") {
			if val.Interface().(string) == "" {
				continue
			}
			if block.Metadata == nil {
				block.Metadata = &Metadata{}
			}
			block.Metadata.Parent = val.Interface().(string)
			continue
		}

		if tagOpts.Has("type") {
			block.Type = val.Interface().(string)
			continue
		}

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Has("omitempty") {
			zero := reflect.Zero(val.Type()).Interface()
			current := val.Interface()

			if reflect.DeepEqual(current, zero) {
				continue
			}
		}

		// if !tagOpts.Has("omitnested") {
		finalVal = s.nested(val)

		v := reflect.ValueOf(val.Interface())
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// switch v.Kind() {
		// case reflect.Map, reflect.Struct:
		// 	isSubStruct = true
		// }
		// } else {
		// 	finalVal = val.Interface()
		// }

		// if tagOpts.Has("string") {
		// 	s, ok := val.Interface().(fmt.Stringer)
		// 	if ok {
		// 		out[name] = s.String()
		// 	}
		// 	continue
		// }

		if v.IsValid() == false {
			continue
		}

		if reflect.PtrTo(v.Type()).Implements(marshalerType) {
			var m Marshaler
			var ok bool
			if v.Kind() == reflect.Ptr {
				m, ok = val.Elem().Interface().(Marshaler)
				if !ok {
					continue
				}
			} else {
				m, ok = val.Interface().(Marshaler)
				if !ok {
					panic("nested non-ptr marshallers not supported for now")
					continue
				}
			}
			b, err := m.MarshalBlock()
			if err != nil {
				panic(err)
			}
			out[name] = b
			continue
		}

		// if isSubStruct && (tagOpts.Has("flatten")) {
		// 	for k := range finalVal.(map[string]interface{}) {
		// 		out[k] = finalVal.(map[string]interface{})[k]
		// 	}
		// } else {
		out[name] = finalVal
		// }
	}

	if block.Type == "" {
		// panic("could not find type for " + reflect.TypeOf(s.raw).String())
	}
}

// structFields returns the exported struct fields for a given s struct. This
// is a convenient helper method to avoid duplicate code in some of the
// functions.
func (s *Struct) structFields() []reflect.StructField {
	t := s.value.Type()

	var f []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// we can't access the value of unexported fields
		if field.PkgPath != "" {
			continue
		}

		// don't check if it's omitted
		if tag := field.Tag.Get(s.TagName); tag == "-" {
			continue
		}

		f = append(f, field)
	}

	return f
}

// nested retrieves recursively all types for the given value and returns the
// nested value.
func (s *Struct) nested(val reflect.Value) interface{} {
	var finalVal interface{}

	v := reflect.ValueOf(val.Interface())
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		n := New(val.Interface())
		n.TagName = s.TagName
		m := n.Map()

		// do not add the converted value if there are no exported fields, ie:
		// time.Time
		if len(m) == 0 {
			finalVal = val.Interface()
		} else {
			finalVal = m
		}
	case reflect.Map:
		// get the element type of the map
		mapElem := val.Type()
		switch val.Type().Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map,
			reflect.Slice, reflect.Chan:
			mapElem = val.Type().Elem()
			if mapElem.Kind() == reflect.Ptr {
				mapElem = mapElem.Elem()
			}
		}

		// only iterate over struct types, ie: map[string]StructType,
		// map[string][]StructType,
		if mapElem.Kind() == reflect.Struct ||
			(mapElem.Kind() == reflect.Slice &&
				mapElem.Elem().Kind() == reflect.Struct) {
			m := make(map[string]interface{}, val.Len())
			for _, k := range val.MapKeys() {
				m[k.String()] = s.nested(val.MapIndex(k))
			}
			finalVal = m
			break
		}

		// TODO(arslan): should this be optional?
		finalVal = val.Interface()
	case reflect.Slice, reflect.Array:
		if val.Type().Kind() == reflect.Interface {
			finalVal = val.Interface()
			break
		}

		// TODO(arslan): should this be optional?
		// do not iterate of non struct types, just pass the value. Ie: []int,
		// []string, co... We only iterate further if it's a struct.
		// i.e []foo or []*foo
		if val.Type().Elem().Kind() != reflect.Struct &&
			!(val.Type().Elem().Kind() == reflect.Ptr &&
				val.Type().Elem().Elem().Kind() == reflect.Struct) {
			finalVal = val.Interface()
			break
		}

		slices := make([]interface{}, val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = s.nested(val.Index(x))
		}
		finalVal = slices
	default:
		finalVal = val.Interface()
	}

	return finalVal
}
