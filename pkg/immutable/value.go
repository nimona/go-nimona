package immutable

import (
	"fmt"
	"strings"
)

type (
	Value interface {
		typeHint() string

		Primitive() interface{}
		PrimitiveHinted() interface{}
	}
	boolValue   struct{ value bool }
	stringValue struct{ value string }
	intValue    struct{ value int64 }
	floatValue  struct{ value float64 }
	bytesValue  struct{ value []byte }
)

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
func (v Map) typeHint() string         { return mapTypeHint }

func (v boolValue) Primitive() interface{}   { return v.value }
func (v stringValue) Primitive() interface{} { return v.value }
func (v intValue) Primitive() interface{}    { return v.value }
func (v floatValue) Primitive() interface{}  { return v.value }
func (v bytesValue) Primitive() interface{}  { return v.value }

func (v boolValue) PrimitiveHinted() interface{}   { return v.value }
func (v stringValue) PrimitiveHinted() interface{} { return v.value }
func (v intValue) PrimitiveHinted() interface{}    { return v.value }
func (v floatValue) PrimitiveHinted() interface{}  { return v.value }
func (v bytesValue) PrimitiveHinted() interface{}  { return v.value }

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
			return boolValue{v}
		}

	case stringTypeHint:
		switch v := a.(type) {
		case string:
			return stringValue{v}
		}

	case intTypeHint:
		switch v := a.(type) {
		case int:
			return intValue{int64(v)}
		}

	case floatTypeHint:
		switch v := a.(type) {
		case float32:
			return floatValue{float64(v)}
		}

	case bytesTypeHint:
		switch v := a.(type) {
		case []byte:
			return bytesValue{v}
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
			return Map{m}
		}

	case listTypeHint:
		switch v := a.(type) {
		case []interface{}:
			m := List{
				hint: strings.Join(hs, ""),
			}
			h := fmt.Sprintf(":%s", strings.Join(hs[1:], ""))
			for _, v := range v {
				m = m.Append(AnyToValue(h, v))
			}
			return m
		}
	}

	// spew.Dump(k, a)
	panic("not sure how to handle")
}
