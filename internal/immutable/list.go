package immutable

type List struct {
	hint  string
	value Value
	prev  *List
}

func (l List) typeHint() string {
	return l.hint
}

func (l List) Append(v Value) List {
	// TODO check type

	if l.value == nil {
		return List{
			hint:  l.hint,
			value: v,
		}
	}

	return List{
		hint:  l.hint,
		value: v,
		prev:  &l,
	}
}

func (l List) Iterate(f func(v Value)) {
	if l.prev != nil {
		l.prev.Iterate(f)
	}
	if l.value != nil {
		f(l.value)
	}
}

func (l List) Length() (n int) {
	l.Iterate(func(v Value) {
		n++
	})
	return
}

func (l List) Primitive() interface{} {
	p := []interface{}{}
	l.Iterate(func(v Value) {
		p = append(p, v.Primitive())
	})
	return p
}

func (l List) PrimitiveHinted() interface{} {
	return l.Primitive()
}
