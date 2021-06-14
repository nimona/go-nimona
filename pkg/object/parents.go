package object

import "nimona.io/pkg/chore"

type (
	Parents map[string]chore.CIDArray
)

func (p Parents) All() []chore.CID {
	m := map[chore.CID]struct{}{}
	for _, ps := range p {
		for _, p := range ps {
			m[p] = struct{}{}
		}
	}
	a := []chore.CID{}
	for c := range m {
		a = append(a, c)
	}
	return a
}
