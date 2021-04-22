package object

type (
	Parents map[string][]CID
)

func (p Parents) All() []CID {
	m := map[CID]struct{}{}
	for _, ps := range p {
		for _, p := range ps {
			m[p] = struct{}{}
		}
	}
	a := []CID{}
	for c := range m {
		a = append(a, c)
	}
	return a
}
