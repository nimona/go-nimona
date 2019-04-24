package immutable

type mapIterator interface {
	Value(k string) interface{}
	Iterate(func(k string, v interface{}))
}

type Map struct {
	mapIterator
}

type empty struct{}

func (empty) Value(_ string) interface{} {
	return nil
}

func (empty) Iterate(f func(k string, v interface{})) {
	// nothing to call
}

func NewMap() Map {
	return Map{empty{}}
}

type pair struct {
	k      string
	v      interface{}
	parent mapIterator
}

func (p pair) Value(k string) interface{} {
	if k == p.k {
		return p.v
	}
	return p.parent.Value(k)
}

func (p pair) Iterate(f func(k string, v interface{})) {
	f(p.k, p.v)
	if p.parent == nil {
		return
	}
	p.parent.Iterate(f)
}

func (m Map) Iterate(f func(k string, v interface{})) {
	if m.mapIterator == nil {
		return
	}
	seen := make(map[interface{}]bool)
	m.mapIterator.Iterate(func(k string, v interface{}) {
		if !seen[k] {
			f(k, v)
			seen[k] = true
		}
	})
}

func (m Map) Value(k string) interface{} {
	if m.mapIterator == nil {
		return nil
	}
	return m.mapIterator.Value(k)
}

func (m Map) Set(k string, v interface{}) Map {
	return Map{pair{k, v, m.mapIterator}}
}
