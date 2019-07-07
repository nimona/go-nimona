package immutable

import (
	"fmt"
	"strings"
)

type (
	typeHinted interface {
		typeHint() string
		primitive() interface{}
	}
	Value struct {
		kind typeHinted
	}
	boolValue   struct{ value bool }
	stringValue struct{ value string }
	intValue    struct{ value int64 }
	floatValue  struct{ value float64 }
	bytesValue  struct{ value []byte }
	mapValue    struct{ value Map }
	listValue   struct {
		hint  string
		value List
	}
)

func (v Value) BoolValue() bool {
	if x, ok := v.kind.(boolValue); ok {
		return x.value
	}
	return false
}
func (v Value) StringValue() string {
	if x, ok := v.kind.(stringValue); ok {
		return x.value
	}
	return ""
}

func (v Value) IntValue() int64 {
	if x, ok := v.kind.(intValue); ok {
		return x.value
	}
	return 0
}

func (v Value) FloatValue() float64 {
	if x, ok := v.kind.(floatValue); ok {
		return x.value
	}
	return 0
}

func (v Value) BytesValue() []byte {
	if x, ok := v.kind.(bytesValue); ok {
		return x.value
	}
	return nil
}

func (v Value) MapValue() Map {
	if x, ok := v.kind.(mapValue); ok {
		return x.value
	}
	return Map{}
}

func (v Value) ListValue() List {
	if x, ok := v.kind.(listValue); ok {
		return x.value
	}
	return List{}
}

func (v Value) Primitive() interface{} {
	switch b := v.kind.(type) {
	case boolValue:
		return v.kind.primitive()
	case stringValue:
		return v.kind.primitive()
	case intValue:
		return v.kind.primitive()
	case floatValue:
		return v.kind.primitive()
	case bytesValue:
		return v.kind.primitive()
	case mapValue:
		return b.value.primitive()
	case listValue:
		return b.value.primitive()
	}
	return nil
}

func (v Value) raw() interface{} {
	switch v.kind.(type) {
	case boolValue:
		return v.BoolValue()
	case stringValue:
		return v.StringValue()
	case intValue:
		return v.IntValue()
	case floatValue:
		return v.FloatValue()
	case bytesValue:
		return v.BytesValue()
	case mapValue:
		return v.MapValue()
	case listValue:
		return v.ListValue()
	}
	return nil
}

func (v Value) PrimitiveHinted() interface{} {
	switch b := v.kind.(type) {
	case boolValue:
		return v.kind.primitive()
	case stringValue:
		return v.kind.primitive()
	case intValue:
		return v.kind.primitive()
	case floatValue:
		return v.kind.primitive()
	case bytesValue:
		return v.kind.primitive()
	case mapValue:
		return b.value.primitiveHinted()
	case listValue:
		return b.value.primitive()
	}
	return nil
}

const (
	boolTypeHint   = "b"
	stringTypeHint = "s"
	intTypeHint    = "i"
	floatTypeHint  = "f"
	bytesTypeHint  = "d"
	mapTypeHint    = "o"
	listTypeHint   = "a"
)

func (v boolValue) typeHint() string   { return boolTypeHint }
func (v stringValue) typeHint() string { return stringTypeHint }
func (v intValue) typeHint() string    { return intTypeHint }
func (v floatValue) typeHint() string  { return floatTypeHint }
func (v bytesValue) typeHint() string  { return bytesTypeHint }
func (v mapValue) typeHint() string    { return mapTypeHint }
func (v listValue) typeHint() string   { return v.hint }

func (v boolValue) primitive() interface{}   { return v.value }
func (v stringValue) primitive() interface{} { return v.value }
func (v intValue) primitive() interface{}    { return v.value }
func (v floatValue) primitive() interface{}  { return v.value }
func (v bytesValue) primitive() interface{}  { return v.value }
func (v mapValue) primitive() interface{}    { return v.value }
func (v listValue) primitive() interface{}   { return v.value }

func getHints(k string) []string {
	ps := strings.Split(k, ":")
	if len(ps) == 1 {
		return nil
	}
	hs := []string{}
	for _, sh := range ps[1] {
		hs = append(hs, string(sh))
	}
	return hs
}

func rmHints(k string) string {
	ps := strings.Split(k, ":")
	if len(ps) == 1 {
		return k
	}
	return ps[0]
}

func AnyToValue(k string, a interface{}) Value {
	hs := getHints(k)

	switch hs[0] {
	case boolTypeHint:
		switch v := a.(type) {
		case bool:
			return Value{
				boolValue{v},
			}
		}

	case stringTypeHint:
		switch v := a.(type) {
		case string:
			return Value{
				stringValue{v},
			}
		}

	case intTypeHint:
		switch v := a.(type) {
		case int:
			return Value{
				intValue{int64(v)},
			}
		}

	case floatTypeHint:
		switch v := a.(type) {
		case float32:
			return Value{
				floatValue{float64(v)},
			}
		}

	case bytesTypeHint:
		switch v := a.(type) {
		case []byte:
			return Value{
				bytesValue{v},
			}
		}

	case mapTypeHint:
		switch v := a.(type) {
		case map[interface{}]interface{}:
			m := Map{}
			for k, v := range v {
				s, ok := k.(string)
				if !ok {
					panic("only string keys are allowed")
				}
				// TODO should we not be checking the hint?
				m = m.Set(rmHints(s), AnyToValue(s, v))
			}
			return Value{
				mapValue{m},
			}
		}

	case listTypeHint:
		switch v := a.(type) {
		case []interface{}:
			m := List{}
			h := fmt.Sprintf(":%s", strings.Join(hs[1:], ""))
			for _, v := range v {
				m = m.Append(AnyToValue(h, v))
			}
			return Value{
				listValue{
					hint:  strings.Join(hs, ""),
					value: m,
				},
			}
		}
	}

	// spew.Dump(k, a)
	panic("not sure how to handle")
}
