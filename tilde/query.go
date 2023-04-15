package tilde

import (
	"errors"
	"regexp"
	"strings"
)

type PathValue interface {
	Value
	getChild(key string) (Value, error)
}

type Query struct {
	m     Map
	path  string
	where func(Value) bool
}

func (m Map) Query() *Query {
	return &Query{m: m}
}

func (q *Query) Select(path string) *Query {
	q.path = path
	return q
}

func (q *Query) Where(cond func(Value) bool) *Query {
	q.where = cond
	return q
}

func (q *Query) Exec() (Value, error) {
	value, err := q.m.getByPath(q.path)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case List:
		result := List{}
		for _, item := range v {
			if q.where == nil || q.where(item) {
				result = append(result, item)
			}
		}
		return result, nil
	case Map:
		if q.where == nil || q.where(value) {
			return value, nil
		}
		return nil, nil
	default:
		return value, nil
	}
}

func (m Map) getByPath(path string) (Value, error) {
	if path == "." || path == "" {
		return m, nil
	}

	parts := strings.Split(path, ".")
	var value Value = m

	for _, part := range parts {
		var err error
		if pathValue, ok := value.(PathValue); ok {
			value, err = pathValue.getChild(part)
		} else {
			return nil, errors.New("value type does not support getChild")
		}
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

func (m Map) getChild(key string) (Value, error) {
	value, ok := m[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (l List) getChild(_ string) (Value, error) {
	return nil, errors.New("cannot get child from List")
}

func Eq(path string, value Value) func(Value) bool {
	return func(v Value) bool {
		if pathValue, ok := v.(PathValue); ok {
			child, err := pathValue.getChild(path)
			if err != nil {
				return false
			}
			return child == value
		}
		return false
	}
}

func Gt(path string, value Value) func(Value) bool {
	return func(v Value) bool {
		if pathValue, ok := v.(PathValue); ok {
			child, err := pathValue.getChild(path)
			if err != nil {
				return false
			}
			switch a := child.(type) {
			case String:
				if b, ok := value.(String); ok {
					return a > b
				}
			case Int64:
				if b, ok := value.(Int64); ok {
					return a > b
				}
			case Uint64:
				if b, ok := value.(Uint64); ok {
					return a > b
				}
			}
		}
		return false
	}
}

func Lt(path string, value Value) func(Value) bool {
	return func(v Value) bool {
		if pathValue, ok := v.(PathValue); ok {
			child, err := pathValue.getChild(path)
			if err != nil {
				return false
			}
			switch a := child.(type) {
			case String:
				if b, ok := value.(String); ok {
					return a < b
				}
			case Int64:
				if b, ok := value.(Int64); ok {
					return a < b
				}
			case Uint64:
				if b, ok := value.(Uint64); ok {
					return a < b
				}
			}
		}
		return false
	}
}

func Like(path, pattern string) func(Value) bool {
	return func(v Value) bool {
		if pathValue, ok := v.(PathValue); ok {
			child, err := pathValue.getChild(path)
			if err != nil {
				return false
			}

			if s, ok := child.(String); ok {
				matched, err := pathMatchesPattern(string(s), pattern)
				if err != nil {
					return false
				}
				return matched
			}
		}
		return false
	}
}

func pathMatchesPattern(s, pattern string) (bool, error) {
	regexPattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), "%", ".*") + "$"
	matched, err := regexp.MatchString(regexPattern, s)
	return matched, err
}
