package object

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
)

type (
	Value interface {
		typeHint() TypeHint

		PrimitiveHinted() interface{}
		Hash() Hash

		IsList() bool
		IsMap() bool
		IsBool() bool
		IsString() bool
		IsRef() bool
		IsInt() bool
		IsFloat() bool
		IsBytes() bool
	}
	Bool   bool
	String string
	Ref    Hash
	Int    int64
	Float  float64
	Bytes  []byte
)

func (v Bool) typeHint() TypeHint   { return HintBool }
func (v String) typeHint() TypeHint { return HintString }
func (v Ref) typeHint() TypeHint    { return HintRef }
func (v Int) typeHint() TypeHint    { return HintInt }
func (v Float) typeHint() TypeHint  { return HintFloat }
func (v Bytes) typeHint() TypeHint  { return HintData }
func (v Map) typeHint() TypeHint    { return HintMap }

func (v Bool) PrimitiveHinted() interface{}   { return bool(v) }
func (v String) PrimitiveHinted() interface{} { return string(v) }
func (v Ref) PrimitiveHinted() interface{}    { return string(v) }
func (v Int) PrimitiveHinted() interface{}    { return int64(v) }
func (v Float) PrimitiveHinted() interface{}  { return float64(v) }
func (v Bytes) PrimitiveHinted() interface{}  { return []byte(v) }

func getHints(k string) typeHints {
	ps := strings.Split(k, ":")
	if len(ps) == 1 {
		return nil
	}
	hs := typeHints{}
	for _, sh := range ps[1] {
		hs = append(hs, TypeHint(sh))
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

// nolint: gocritic
func AnyToValue(k string, a interface{}) Value {
	hs := getHints(k)

	if a == nil {
		return nil
	}

	if len(hs) == 0 {
		panic("missing hints; k=" + k)
	}

	switch hs[0] {
	case HintBool:
		switch v := a.(type) {
		case bool:
			return Bool(v)
		}

	case HintRef:
		switch v := a.(type) {
		case string:
			return Ref(v)
		}

	case HintString:
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

	case HintInt:
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
			return Int(v)
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

	case HintFloat:
		switch v := a.(type) {
		case float32:
			return Float(float64(v))
		case float64:
			return Float(v)
		}

	case HintData:
		switch v := a.(type) {
		case []byte:
			return Bytes(v)
		case string:
			b, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				panic(err)
			}
			return Bytes(b)
		}

	case HintMap:
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

	case HintArray:
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
				hint: hs.TypeHint(),
			}
			h := fmt.Sprintf(":%s", hs[1:].TypeHint().String())
			for _, v := range v {
				if v != nil {
					m = m.Append(AnyToValue(h, v))
				}
			}
			return m
		}

		switch reflect.TypeOf(a).Kind() {
		case reflect.Slice:
			h := fmt.Sprintf(":%s", hs[1:].TypeHint().String())
			m := List{
				hint: hs.TypeHint(),
			}
			v := reflect.ValueOf(a)
			for i := 0; i < v.Len(); i++ {
				m = m.Append(AnyToValue(h, v.Index(i).Interface()))
			}
			return m
		}
	}

	panic(
		fmt.Sprintf(
			"not sure how to handle; k=%s a=%#v t=%s",
			k, a, reflect.TypeOf(a).String(),
		),
	)
}
