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

func (l List) Iterate(f func(v Value) bool) {
	l.iterate(f)
}

func (l List) Length() (n int) {
	l.Iterate(func(v Value) bool {
		n++
		return true
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
		l.Iterate(func(v Value) bool {
			p = append(p, v.PrimitiveHinted())
			return true
		})
		return p
	case l.value.IsMap():
		p := []interface{}{}
		l.Iterate(func(v Value) bool {
			p = append(p, v.PrimitiveHinted())
			return true
		})
		return p
	case l.value.IsBool():
		p := []bool{}
		l.Iterate(func(v Value) bool {
			p = append(p, v.PrimitiveHinted().(bool))
			return true
		})
		return p
	case l.value.IsString():
		p := []string{}
		l.Iterate(func(v Value) bool {
			p = append(p, v.PrimitiveHinted().(string))
			return true
		})
		return p
	case l.value.IsInt():
		p := []int64{}
		l.Iterate(func(v Value) bool {
			p = append(p, v.PrimitiveHinted().(int64))
			return true
		})
		return p
	case l.value.IsFloat():
		p := []float64{}
		l.Iterate(func(v Value) bool {
			p = append(p, v.PrimitiveHinted().(float64))
			return true
		})
		return p
	case l.value.IsBytes():
		p := [][]byte{}
		l.Iterate(func(v Value) bool {
			p = append(p, v.PrimitiveHinted().([]byte))
			return true
		})
		return p
	}

	return nil
}
