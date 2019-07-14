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

func (v Bool) Primitive() interface{}   { return bool(v) }
func (v String) Primitive() interface{} { return string(v) }
func (v Int) Primitive() interface{}    { return int64(v) }
func (v Float) Primitive() interface{}  { return float64(v) }
func (v Bytes) Primitive() interface{}  { return []byte(v) }

func (v Bool) PrimitiveHinted() interface{}   { return v.Primitive() }
func (v String) PrimitiveHinted() interface{} { return v.Primitive() }
func (v Int) PrimitiveHinted() interface{}    { return v.Primitive() }
func (v Float) PrimitiveHinted() interface{}  { return v.Primitive() }
func (v Bytes) PrimitiveHinted() interface{}  { return v.Primitive() }

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
			return Bool(v)
		}

	case stringTypeHint:
		switch v := a.(type) {
		case string:
			return String(v)
		}

	case intTypeHint:
		switch v := a.(type) {
		case int:
			return Int(int64(v))
		}

	case floatTypeHint:
		switch v := a.(type) {
		case float32:
			return Float(float64(v))
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
