package immutable

type List struct {
	listList
}

type listList interface {
	Value() Value
	Iterate(f func(v Value))
	Length() (n int)
	Append(v Value) List
}

type listIterator struct {
	value Value
	prev  List
}

func (l listIterator) Value() Value {
	return l.value
}

func (l listIterator) Previous() List {
	return l.prev
}

func (l List) Append(v Value) List {
	if l.listList == nil {
		return List{
			listIterator{
				value: v,
			},
		}
	}

	return l.listList.Append(v)
}

func (l listIterator) Append(v Value) List {
	return List{
		listIterator{
			value: v,
			prev:  List{l},
		},
	}
}

func (l List) Iterate(f func(v Value)) {
	if l.listList == nil {
		return
	}

	l.listList.Iterate(f)
}

func (l listIterator) Iterate(f func(v Value)) {
	f(l.value)
	if l.prev.listList == nil {
		return
	}
	l.prev.Iterate(f)
}

func (l List) Length() int {
	if l.listList == nil {
		return 0
	}

	return l.listList.Length()
}

func (l listIterator) Length() (n int) {
	l.Iterate(func(v Value) {
		n++
	})
	return
}

func (l List) primitive() interface{} {
	p := []interface{}{}
	l.Iterate(func(v Value) {
		p = append(p, v.raw())
	})
	return p
}
