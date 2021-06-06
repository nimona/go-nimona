package object

import "nimona.io/pkg/object/value"

type (
	Parents map[string]value.CIDArray
)

func (p Parents) All() []value.CID {
	m := map[value.CID]struct{}{}
	for _, ps := range p {
		for _, p := range ps {
			m[p] = struct{}{}
		}
	}
	a := []value.CID{}
	for c := range m {
		a = append(a, c)
	}
	return a
}
