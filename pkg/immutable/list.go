package immutable

import (
	"github.com/cheekybits/genny/generic"
)

type ValueTypeList struct {
	value []generic.Type
	// listList
}

// type listList interface {
// 	Value() Value
// 	Next() List
// }

// type listItem struct {
// 	value Value
// 	next  List
// }

// func (n listItem) Value() generic.Type {
// 	return n.value
// }

// func (n listItem) Next() List {
// 	return n.next
// }

func (l ValueTypeList) Append(v generic.Type) ValueTypeList {
	nv := make([]generic.Type, len(l.value), len(l.value)+1)
	copy(nv, l.value)
	return ValueTypeList{value: append(nv, v)}
}
