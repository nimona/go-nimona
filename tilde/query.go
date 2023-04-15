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
	m          Value
	path       string
	conditions []func(Value) bool
}

func (m Map) Query() *Query {
	return &Query{m: m}
}

func (l List) Query() *Query {
	return &Query{m: l}
}

func (q *Query) Select(path string) *Query {
	q.path = path
	return q
}

func (q *Query) Where(cond ...func(Value) bool) *Query {
	q.conditions = append(q.conditions, cond...)
	return q
}

func (q *Query) Exec() (Value, error) {
	var value Value
	var err error
	switch v := q.m.(type) {
	case Map:
		value, err = v.getByPath(q.path)
	case List:
		value, err = v.getByPath(q.path)
	default:
		return nil, errors.New("unsupported value type")
	}
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case List:
		result := List{}
		for _, item := range v {
			shouldInclude := true
			for _, cond := range q.conditions {
				if !cond(item) {
					shouldInclude = false
					break
				}
			}
			if shouldInclude {
				result = append(result, item)
			}
		}
		return result, nil
	case Map:
		shouldInclude := true
		for _, cond := range q.conditions {
			if !cond(value) {
				shouldInclude = false
				break
			}
		}
		if shouldInclude {
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

func (l List) getByPath(path string) (Value, error) {
	subResults := List{}

	for _, item := range l {
		switch v := item.(type) {
		case Map:
			subResult, err := v.getByPath(path)
			if err != nil {
				return nil, err
			}
			switch subResult := subResult.(type) {
			case List:
				subResults = append(subResults, subResult...)
			default:
				subResults = append(subResults, subResult)
			}
		case List:
			subResult, err := v.getByPath(path)
			if err != nil {
				return nil, err
			}
			subResults = append(subResults, subResult.(List)...)
		default:
			return nil, errors.New("unsupported value type")
		}
	}

	return subResults, nil
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
			r, err := child.cmp(value)
			if err != nil {
				return false
			}
			return r == 0
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
			r, err := child.cmp(value)
			if err != nil {
				return false
			}
			return r == 1
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
			r, err := child.cmp(value)
			if err != nil {
				return false
			}
			return r == -1
		}
		return false
	}
}

func Gte(path string, value Value) func(Value) bool {
	return func(v Value) bool {
		if pathValue, ok := v.(PathValue); ok {
			child, err := pathValue.getChild(path)
			if err != nil {
				return false
			}
			r, err := child.cmp(value)
			if err != nil {
				return false
			}
			return r >= 0
		}
		return false
	}
}

func Lte(path string, value Value) func(Value) bool {
	return func(v Value) bool {
		if pathValue, ok := v.(PathValue); ok {
			child, err := pathValue.getChild(path)
			if err != nil {
				return false
			}
			r, err := child.cmp(value)
			if err != nil {
				return false
			}
			return r <= 0
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
