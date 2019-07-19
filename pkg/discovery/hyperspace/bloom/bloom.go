package bloom

import (
	"sort"
)

type (
	// Bloom -
	Bloom []int
	// Bloomer -
	Bloomer interface {
		Bloom() []int
	}
)

func (b Bloom) Bloom() []int {
	return []int(b)
}

func NewBloom(contentHashes ...string) Bloom {
	bs := []int{}
	for _, c := range contentHashes {
		b := HashChunked("", []byte(c))
		bs = append(bs, b...)
	}
	bs = unique(bs)
	sort.Ints(bs)
	return Bloom(bs)
}

func unique(s []int) []int {
	seen := make(map[int]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}
