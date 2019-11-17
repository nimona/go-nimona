package bloom

import (
	"sort"
)

type (
	// Bloom -
	Bloom []int64
	// Bloomer -
	Bloomer interface {
		Bloom() []int64
	}
)

func (b Bloom) Bloom() []int64 {
	return []int64(b)
}

func New(contentHashes ...string) Bloom {
	bs := []int{}
	for _, c := range contentHashes {
		b := HashChunked("", []byte(c))
		bs = append(bs, b...)
	}
	bs = unique(bs)
	sort.Ints(bs)
	bs64 := make([]int64, len(bs))
	for i, v := range bs {
		bs64[i] = int64(v)
	}
	return Bloom(bs64)
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
