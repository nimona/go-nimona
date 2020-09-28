package object

type List struct {
	hint  TypeHint
	index int
	size  int
	value Value
	prev  *List
}

func (l List) typeHint() TypeHint {
	return l.hint
}

func (l List) set(index int, v Value) List {
	// TODO check type
	cp := l.size
	if index > cp {
		cp = index + 1
	}
	var prev *List
	if l.value != nil {
		prev = &l
	}
	hint := l.hint
	if hint == "" {
		hint = "a" + v.typeHint()
	}
	return List{
		hint:  hint,
		index: index,
		size:  cp,
		value: v,
		prev:  prev,
	}
}

func (l List) Set(path string, v Value) List {
	nm, err := setPath(l, path, v)
	if err != nil {
		panic(err)
	}
	return nm.(List)
}

func (l List) Append(v Value) List {
	// TODO check type
	if l.value == nil {
		return List{
			hint:  "a" + v.typeHint(),
			index: 0,
			size:  1,
			value: v,
		}
	}

	return List{
		hint:  l.hint,
		index: l.size,
		size:  l.size + 1,
		value: v,
		prev:  &l,
	}
}

func (l List) iterate(f func(v Value) bool) bool {
	if l.prev != nil {
		if !l.prev.iterate(f) {
			return false
		}
	}
	if l.value != nil {
		if !f(l.value) {
			return false
		}
	}
	return true
}

// nolint: unused // we'll probably needs this again
func (l List) iterateRev(f func(v Value) bool) bool {
	if l.value != nil {
		if !f(l.value) {
			return false
		}
	}
	if l.prev != nil {
		if !l.prev.iterate(f) {
			return false
		}
	}
	return true
}

func (l List) Get(index int) Value {
	ll := &l
	for {
		if ll == nil {
			return nil
		}
		if ll.index == index {
			return ll.value
		}
		ll = ll.prev
	}
}

func (l List) Iterate(f func(int, Value) bool) {
	if l.size == 0 {
		return
	}
	vs := make([]*Value, l.size)
	ll := &l
	for {
		if ll == nil {
			break
		}
		if vs[ll.index] == nil {
			vs[ll.index] = &ll.value
		}
		ll = ll.prev
	}
	for i, iv := range vs {
		if iv == nil {
			continue
		}
		if !f(i, *iv) {
			break
		}
	}
}

func (l List) Length() (n int) {
	// TODO this is wrong due to the way set works
	return l.size
}

func (l List) PrimitiveHinted() interface{} {
	if l.value == nil {
		if l.prev == nil {
			return nil
		}
		return l.prev.PrimitiveHinted()
	}

	// TODO should we be keeping the original indices? ie:
	// > p := make([]interface{}, l.size)
	// > p[i] = v.PrimitiveHinted()

	switch {
	case l.value.IsList():
		p := []interface{}{}
		l.Iterate(func(i int, v Value) bool {
			p = append(p, v.PrimitiveHinted())
			return true
		})
		return p
	case l.value.IsMap():
		p := []interface{}{}
		l.Iterate(func(i int, v Value) bool {
			p = append(p, v.PrimitiveHinted())
			return true
		})
		return p
	case l.value.IsBool():
		p := []bool{}
		l.Iterate(func(i int, v Value) bool {
			p = append(p, v.PrimitiveHinted().(bool))
			return true
		})
		return p
	case l.value.IsString():
		p := []string{}
		l.Iterate(func(i int, v Value) bool {
			p = append(p, v.PrimitiveHinted().(string))
			return true
		})
		return p
	case l.value.IsInt():
		p := []int64{}
		l.Iterate(func(i int, v Value) bool {
			p = append(p, v.PrimitiveHinted().(int64))
			return true
		})
		return p
	case l.value.IsFloat():
		p := []float64{}
		l.Iterate(func(i int, v Value) bool {
			p = append(p, v.PrimitiveHinted().(float64))
			return true
		})
		return p
	case l.value.IsBytes():
		p := [][]byte{}
		l.Iterate(func(i int, v Value) bool {
			p = append(p, v.PrimitiveHinted().([]byte))
			return true
		})
		return p
	}

	return nil
}
