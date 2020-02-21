package immutable

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	Value interface {
		typeHint() string

		PrimitiveHinted() interface{}

		IsList() bool
		IsMap() bool
		IsBool() bool
		IsString() bool
		IsInt() bool
		IsFloat() bool
		IsBytes() bool
	}
	Bool   bool
	String string
	Int    int64
	Float  float64
	Bytes  []byte
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

func (v Bool) typeHint() string   { return boolTypeHint }
func (v String) typeHint() string { return stringTypeHint }
func (v Int) typeHint() string    { return intTypeHint }
func (v Float) typeHint() string  { return floatTypeHint }
func (v Bytes) typeHint() string  { return bytesTypeHint }
func (v Map) typeHint() string    { return mapTypeHint }

func (v Bool) PrimitiveHinted() interface{}   { return bool(v) }
func (v String) PrimitiveHinted() interface{} { return string(v) }
func (v Int) PrimitiveHinted() interface{}    { return int64(v) }
func (v Float) PrimitiveHinted() interface{}  { return float64(v) }
func (v Bytes) PrimitiveHinted() interface{}  { return []byte(v) }

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
	// ps := strings.Split(k, ":")
	// if len(ps) == 1 {
	return k
	// }
	// return ps[0]
}

func AnyToValue(k string, a interface{}) Value {
	hs := getHints(k)

	if a == nil {
		return nil
	}

	if len(hs) == 0 {
		panic("missing hints; k=" + k)
	}

	switch hs[0] {
	case boolTypeHint:
		switch v := a.(type) {
		case bool:
			return Bool(v)
		}

	case stringTypeHint:
		switch v := a.(type) {
		case string:
			return String(v)
		}

		if s, ok := a.(string); ok {
			return String(s)
		}

		if s, ok := a.(interface{ String() string }); ok {
			return String(s.String())
		}

	case intTypeHint:
		switch v := a.(type) {
		case int:
			return Int(int64(v))
		case int8:
			return Int(int64(v))
		case int16:
			return Int(int64(v))
		case int32:
			return Int(int64(v))
		case int64:
			return Int(int64(v))
		case uint:
			return Int(int64(v))
		case uint8:
			return Int(int64(v))
		case uint16:
			return Int(int64(v))
		case uint32:
			return Int(int64(v))
		case uint64:
			return Int(int64(v))
		}

	case floatTypeHint:
		switch v := a.(type) {
		case float32:
			return Float(float64(v))
		case float64:
			return Float(v)
		}

	case bytesTypeHint:
		switch v := a.(type) {
		case []byte:
			return Bytes(v)
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
				if v != nil {
					// TODO should we not be checking the hint?
					m = m.Set(rmHints(s), AnyToValue(s, v))
				}
			}
			return Map{m: m.m}
		case map[string]interface{}:
			m := Map{}
			for k, v := range v {
				if v != nil {
					m = m.Set(rmHints(k), AnyToValue(k, v))
				}
			}
			return Map{m: m.m}
		}

	case listTypeHint:
		switch v := a.(type) {
		case []string:
			m := List{
				hint: "as",
			}
			for _, v := range v {
				m = m.Append(String(v))
			}
			return m
		case []bool:
			m := List{
				hint: "ab",
			}
			for _, v := range v {
				m = m.Append(Bool(v))
			}
			return m
		case []int64:
			m := List{
				hint: "ai",
			}
			for _, v := range v {
				m = m.Append(Int(v))
			}
			return m
		case []float64:
			m := List{
				hint: "af",
			}
			for _, v := range v {
				m = m.Append(Float(v))
			}
			return m
		case [][]byte:
			m := List{
				hint: "ad",
			}
			for _, v := range v {
				m = m.Append(Bytes(v))
			}
			return m
		case []interface{}:
			m := List{
				hint: strings.Join(hs, ""),
			}
			h := fmt.Sprintf(":%s", strings.Join(hs[1:], ""))
			for _, v := range v {
				if v != nil {
					m = m.Append(AnyToValue(h, v))
				}
			}
			return m
		}

		switch reflect.TypeOf(a).Kind() {
		case reflect.Slice:
			h := fmt.Sprintf(":%s", strings.Join(hs[1:], ""))
			m := List{
				hint: strings.Join(hs, ""),
			}
			v := reflect.ValueOf(a)
			for i := 0; i < v.Len(); i++ {
				m = m.Append(AnyToValue(h, v.Index(i).Interface()))
			}
			return m
		}
	}

	panic(fmt.Sprintf("not sure how to handle; k=%s a=%#v t=%s", k, a, reflect.TypeOf(a).String()))
}
