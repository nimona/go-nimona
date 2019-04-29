package immutable

type Map struct {
	mapIterator
}

type mapIterator interface {
	Value(k string) Value
	Iterate(func(k string, v Value))
}

// type emptyMap struct{}

// func (emptyMap) Value(_ string) Value {
// 	return Value{}
// }

// func (emptyMap) Iterate(f func(k string, v Value)) {
// 	// nothing to call
// }

// func NewMap() Map {
// 	return Map{emptyMap{}}
// }

type mapPair struct {
	k      string
	v      Value
	parent mapIterator
}

func (p mapPair) Value(k string) Value {
	if k == p.k {
		return p.v
	}
	return p.parent.Value(k)
}

func (p mapPair) Iterate(f func(k string, v Value)) {
	f(p.k, p.v)
	if p.parent == nil {
		return
	}
	p.parent.Iterate(f)
}

func (m Map) Iterate(f func(k string, v Value)) {
	if m.mapIterator == nil {
		return
	}
	seen := make(map[string]bool)
	m.mapIterator.Iterate(func(k string, v Value) {
		if !seen[k] {
			f(k, v)
			seen[k] = true
		}
	})
}

func (m Map) Value(k string) Value {
	if m.mapIterator == nil {
		return Value{}
	}
	return m.mapIterator.Value(k)
}

func (m Map) Set(k string, v Value) Map {
	return Map{mapPair{k, v, m.mapIterator}}
}

func (m Map) Primitive() map[string]interface{} {
	p := map[string]interface{}{}
	m.Iterate(func(k string, v Value) {
		p[k] = v.Primitive()
	})
	return p
}

func (m Map) PrimitiveHinted() map[string]interface{} {
	p := map[string]interface{}{}
	m.Iterate(func(k string, v Value) {
		p[k+":"+v.kind.typeHint()] = v.PrimitiveHinted()
	})
	return p
}
