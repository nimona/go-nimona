package immutable

type Map struct {
	m mapIterator
}

type mapIterator interface {
	value(k string) Value
	iterate(func(k string, v Value))
}

type mapPair struct {
	k      string
	v      Value
	parent mapIterator
}

func (p mapPair) value(k string) Value {
	if k == p.k {
		return p.v
	}
	if p.parent == nil {
		return nil
	}
	return p.parent.value(k)
}

func (p mapPair) iterate(f func(k string, v Value)) {
	f(p.k, p.v)
	if p.parent == nil {
		return
	}
	p.parent.iterate(f)
}

func (m Map) Iterate(f func(k string, v Value)) {
	if m.m == nil {
		return
	}
	seen := make(map[string]bool)
	m.m.iterate(func(k string, v Value) {
		if !seen[k] {
			f(k, v)
			seen[k] = true
		}
	})
}

func (m Map) Value(k string) Value {
	// ps := strings.Split(k, ":")
	// if len(ps) > 1 {
	// 	k = ps[0]
	// }
	if m.m == nil {
		return nil
	}
	return m.m.value(k)
}

func (m Map) Set(k string, v Value) Map {
	// ps := strings.Split(k, ":")
	// if len(ps) > 1 {
	// 	k = ps[0]
	// 	// if ps[1] != v.typeHint() {
	// 	// 	panic("hint does not match value type")
	// 	// }
	// }
	return Map{
		mapPair{
			k:      k,
			v:      v,
			parent: m.m,
		},
	}
}

func (m Map) PrimitiveHinted() interface{} {
	if m.m == nil {
		return nil
	}
	p := map[string]interface{}{}
	m.Iterate(func(k string, v Value) {
		// h := ":" + v.typeHint()
		// if strings.HasSuffix(k, h) {
		p[k] = v.PrimitiveHinted()
		// } else {
		// 	p[k+h] = v.PrimitiveHinted()
		// }

	})
	return p
}

func (m Map) IsEmpty() bool {
	return m.m == nil
}
