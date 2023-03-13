package tilde

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/copystructure"
	"github.com/valyala/fastjson"
)

// Document is a map of key/value pairs where keys are strings and values
// are either primitive values, or other documents, as well as lists of
// either of these.
//
// A value can be one of the following kinds:
// - String (string)
// - Int64 (int64)
// - Uint64 (uint64)
// - Bytes ([]byte)
// - Bool (bool)
// - List (List of Values)
// - Map (map of string to Value)
//
// Values can be accessed using the Get and Set methods, and support nesting
// using the dot notation.
//
// Documents are strongly typed and their shape is defined by their schema.
// Trying to access a key that is not defined in the schema will result in an
// error, same goes for trying to set a value that does not match the type
// defined in the schema.
//
// Documents are the basic unit of data in Nimona. They are used to represent
// all other data types.
type (
	ValueKind int
	Hint      rune
	Schema    struct {
		Kind       ValueKind
		Properties map[string]*Schema
		Elements   *Schema
	}
)

const (
	HintInvalid Hint = '?'
	HintString  Hint = 's'
	HintInt64   Hint = 'i'
	HintUint64  Hint = 'u'
	HintBytes   Hint = 'd'
	HintRef     Hint = 'r'
	HintBool    Hint = 'b'
	HintList    Hint = 'a'
	HintMap     Hint = 'm'
	HintAny     Hint = '*'
)

const (
	KindInvalid ValueKind = iota
	KindString  ValueKind = iota
	KindInt64   ValueKind = iota
	KindUint64  ValueKind = iota
	KindBytes   ValueKind = iota
	KindRef     ValueKind = iota
	KindBool    ValueKind = iota
	KindList    ValueKind = iota
	KindMap     ValueKind = iota
	KindAny     ValueKind = iota
)

type (
	Value interface {
		hint() Hint
	}
	xx struct{}

	Int64  int64
	Uint64 uint64
	String string
	Bytes  []byte
	Ref    []byte
	Bool   bool
	Map    map[string]Value
	List   []Value
)

func (Int64) hint() Hint  { return HintInt64 }
func (Uint64) hint() Hint { return HintUint64 }
func (String) hint() Hint { return HintString }
func (Bytes) hint() Hint  { return HintBytes }
func (Bool) hint() Hint   { return HintBool }
func (Map) hint() Hint    { return HintMap }
func (List) hint() Hint   { return HintList }
func (Ref) hint() Hint    { return HintRef }

func (k ValueKind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindInt64:
		return "int64"
	case KindUint64:
		return "uint64"
	case KindBytes:
		return "bytes"
	case KindBool:
		return "bool"
	case KindList:
		return "list"
	case KindMap:
		return "map"
	case KindAny:
		return "Value"
	}
	return "invalid"
}

func (k ValueKind) Name() string {
	switch k {
	case KindString:
		return "String"
	case KindInt64:
		return "Int64"
	case KindUint64:
		return "Uint64"
	case KindBytes:
		return "Bytes"
	case KindBool:
		return "Bool"
	case KindList:
		return "List"
	case KindMap:
		return "Map"
	case KindAny:
		return "Value"
	}
	return "InvalidValueKind" + strconv.Itoa(int(k))
}

func (k ValueKind) Hint() Hint {
	switch k {
	case KindString:
		return HintString
	case KindInt64:
		return HintInt64
	case KindUint64:
		return HintUint64
	case KindBytes:
		return HintBytes
	case KindBool:
		return HintBool
	case KindList:
		return HintList
	case KindMap:
		return HintMap
	case KindAny:
		return HintAny
	}
	return HintInvalid
}

func KindFromString(s string) ValueKind {
	switch s {
	case "string":
		return KindString
	case "int64":
		return KindInt64
	case "uint64":
		return KindUint64
	case "bytes":
		return KindBytes
	case "bool":
		return KindBool
	case "array":
		return KindList
	case "map":
		return KindMap
	}
	return KindInvalid
}

// Get returns the value at the given path.
// Supports dot notation for nested values.
func (m Map) Get(path string) (Value, error) {
	keyFirst, keyRest, _ := strings.Cut(path, ".")
	v, ok := m[keyFirst]
	if !ok {
		return nil, fmt.Errorf("key %s not found", keyFirst)
	}

	if keyRest == "" {
		return v, nil
	}

	switch v := v.(type) {
	case Map:
		return v.Get(keyRest)
	case List:
		return v.Get(keyRest)
	default:
		return v, nil
	}
}

// Get returns the value at the given index.
func (l List) Get(index string) (Value, error) {
	i, keyRest, _ := strings.Cut(index, ".")
	v, err := strconv.Atoi(i)
	if err != nil {
		return nil, err
	}

	if v >= len(l) {
		return nil, fmt.Errorf("index %d out of range", v)
	}

	if keyRest == "" {
		return l[v], nil
	}

	switch v := l[v].(type) {
	case Map:
		return v.Get(keyRest)
	case List:
		return v.Get(keyRest)
	default:
		return nil, fmt.Errorf("index %d is not a map or list", v)
	}
}

// Set sets the value at the given path.
// Supports dot notation for nested values.
func (m Map) Set(path string, value Value) error {
	// attempt to figure out the kind of the value
	// TODO: at some point the map should accept a schema
	keyFirst, keyRest, _ := strings.Cut(path, ".")
	v, ok := m[keyFirst]
	if !ok && strings.Contains(keyRest, ".") {
		keySecond, _, _ := strings.Cut(keyRest, ".")
		if _, err := strconv.Atoi(keySecond); err == nil {
			m[keyFirst] = List{}
			v = m[keyFirst]
		} else {
			m[keyFirst] = Map{}
			v = m[keyFirst]
		}
	}

	if keyRest == "" {
		m[keyFirst] = value
		return nil
	}

	switch v := v.(type) {
	case Map:
		return v.Set(keyRest, value)
	case List:
		return v.Set(keyRest, value)
	default:
		return fmt.Errorf("key %s is not a map or list", keyFirst)
	}
}

// Set sets the value at the given path.
func (l List) Set(path string, value Value) error {
	i, keyRest, _ := strings.Cut(path, ".")
	v, err := strconv.Atoi(i)
	if err != nil {
		return err
	}

	if v >= len(l) {
		return fmt.Errorf("index %d out of range", v)
	}

	if keyRest == "" {
		l[v] = value
		return nil
	}

	switch v := l[v].(type) {
	case Map:
		return v.Set(keyRest, value)
	case List:
		return v.Set(keyRest, value)
	default:
		return fmt.Errorf("index %d is not a map or list", v)
	}
}

// Append appends the given value to the list.
// If the list is nil, a new list is created.
// Supports dot notation for nested values.
func (m Map) Append(path string, value Value) error {
	keyFirst, keyRest, _ := strings.Cut(path, ".")
	v, ok := m[keyFirst]
	if !ok {
		if strings.Contains(keyRest, ".") {
			// if the key rest contains a dot, we need to figure out
			// if the next level is a map or a list
			// TODO: at some point we should use a schema for this
			keySecond, _, _ := strings.Cut(keyRest, ".")
			if _, err := strconv.Atoi(keySecond); err == nil {
				m[keyFirst] = List{}
				v = m[keyFirst]
			} else {
				m[keyFirst] = Map{}
				v = m[keyFirst]
			}
		} else {
			// else we assume it's a list
			m[keyFirst] = List{value}
			return nil
		}
	}

	switch v := v.(type) {
	case Map:
		return v.Append(keyRest, value)
	case List:
		lv, err := v.appendPath(keyRest, value)
		if err != nil {
			return fmt.Errorf("cannot append to list: %v", err)
		}
		m[keyFirst] = lv
		return err
	default:
		return fmt.Errorf("key %s is not a map or list, got %T", keyFirst, v)
	}
}

// appendPath appends the given value to the list.
// If the list is nil, a new list is created.
// Supports dot notation for nested values.
func (l List) appendPath(path string, value Value) (List, error) {
	if path == "" {
		return append(l, value), nil
	}

	keyFirst, keyRest, _ := strings.Cut(path, ".")
	index, err := strconv.Atoi(keyFirst)
	if err != nil {
		return l, fmt.Errorf("invalid index %d", index)
	}

	if keyRest == "" {
		return append(l, value), nil
	}

	if index >= len(l) {
		return l, fmt.Errorf("index %d out of range", index)
	}

	switch v := l[index].(type) {
	case Map:
		err := v.Set(keyRest, value)
		if err != nil {
			return l, fmt.Errorf("error setting %s: %s", keyRest, err)
		}
		l[index] = v
		return l, nil
	case List:
		lv, err := v.appendPath(keyRest, value)
		if err != nil {
			return l, fmt.Errorf("error appending %s: %s", keyRest, err)
		}
		l[index] = lv
		return l, nil
	default:
		return l, fmt.Errorf("index %d is not a map or list", v)
	}
}

func (m *Map) UnmarshalJSON(data []byte) error {
	var sc fastjson.Scanner

	sc.Init(string(data))

	if !sc.Next() {
		return fmt.Errorf("expected object, got nothing")
	}

	o := sc.Value()
	if o.Type() != fastjson.TypeObject {
		return fmt.Errorf("expected object, got %s", o.Type())
	}

	nm, err := unmarshalValue(o)
	if err != nil {
		return fmt.Errorf("error unmarshaling: %w", err)
	}

	*m = nm.(Map)

	return nil
}

func (m Map) MarshalJSON() ([]byte, error) {
	mm := make(map[string]interface{}, len(m))
	for k, v := range m {
		hk := keyWithHint(k, v.hint())
		switch v := v.(type) {
		case Map:
			b, err := v.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("error marshaling map: %w", err)
			}
			mm[hk] = json.RawMessage(b)
		case List:
			cv := v
			hints := []Hint{
				cv.hint(),
			}
			for len(cv) != 0 && cv[0] != nil {
				hints = append(hints, cv[0].hint())
				if cv[0].hint() != HintList {
					break
				}
				cv = cv[0].(List)
			}
			hk = keyWithHint(k, hints...)
			mm[hk] = v
		default:
			mm[hk] = v
		}
	}
	return json.Marshal(mm)
}

func keyWithHint(key string, hints ...Hint) string {
	if len(hints) == 0 {
		return key
	}
	suffix := ""
	for _, hint := range hints {
		suffix += string(hint)
	}
	return fmt.Sprintf("%s:%s", key, suffix)
}

func Copy[T Value](v T) T {
	nv, err := copystructure.Copy(v)
	if err != nil {
		panic(fmt.Errorf("error copying value of type %T: %w", v, err))
	}
	return nv.(T)
}
