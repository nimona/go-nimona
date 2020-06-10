package object

type List struct {
	hint  TypeHint
	value Value
	prev  *List
}

func (l List) typeHint() TypeHint {
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

func (l List) PrimitiveHinted() interface{} {
	if l.value == nil {
		return nil
	}

	switch {
	case l.value.IsList():
		p := []interface{}{}
		l.Iterate(func(v Value) {
			p = append(p, v.PrimitiveHinted())
		})
		return p
	case l.value.IsMap():
		p := []interface{}{}
		l.Iterate(func(v Value) {
			p = append(p, v.PrimitiveHinted())
		})
		return p
	case l.value.IsBool():
		p := []bool{}
		l.Iterate(func(v Value) {
			p = append(p, v.PrimitiveHinted().(bool))
		})
		return p
	case l.value.IsString():
		p := []string{}
		l.Iterate(func(v Value) {
			p = append(p, v.PrimitiveHinted().(string))
		})
		return p
	case l.value.IsInt():
		p := []int64{}
		l.Iterate(func(v Value) {
			p = append(p, v.PrimitiveHinted().(int64))
		})
		return p
	case l.value.IsFloat():
		p := []float64{}
		l.Iterate(func(v Value) {
			p = append(p, v.PrimitiveHinted().(float64))
		})
		return p
	case l.value.IsBytes():
		p := [][]byte{}
		l.Iterate(func(v Value) {
			p = append(p, v.PrimitiveHinted().([]byte))
		})
		return p
	}

	return nil
}
