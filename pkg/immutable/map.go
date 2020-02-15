package immutable

import "strings"

type Map struct {
	mapIterator
}

type mapIterator interface {
	Value(k string) Value
	Iterate(func(k string, v Value))
}

type mapPair struct {
	k      string
	v      Value
	parent mapIterator
}

func (p mapPair) Value(k string) Value {
	if k == p.k {
		return p.v
	}
	if p.parent == nil {
		return nil
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
	ps := strings.Split(k, ":")
	if len(ps) > 1 {
		k = ps[0]
	}
	if m.mapIterator == nil {
		return nil
	}
	return m.mapIterator.Value(k)
}

func (m Map) Set(k string, v Value) Map {
	ps := strings.Split(k, ":")
	if len(ps) > 1 {
		k = ps[0]
		// if ps[1] != v.typeHint() {
		// 	panic("hint does not match value type")
		// }
	}
	return Map{
		mapPair{
			k:      k,
			v:      v,
			parent: m.mapIterator,
		},
	}
}

func (m Map) PrimitiveHinted() interface{} {
	if m.mapIterator == nil {
		return nil
	}
	p := map[string]interface{}{}
	m.Iterate(func(k string, v Value) {
		p[k+":"+v.typeHint()] = v.PrimitiveHinted()
	})
	return p
}
