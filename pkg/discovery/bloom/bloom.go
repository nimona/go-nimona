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

func (b Bloom) Contains(needle []int64) bool {
	return intersectionCount(needle, b.Bloom()) == len(needle)
}

func intersectionCount(a, b []int64) int {
	m := make(map[int64]uint64)
	for _, k := range a {
		m[k] |= (1 << 0)
	}
	for _, k := range b {
		m[k] |= (1 << 1)
	}

	i := 0
	for _, v := range m {
		a := v&(1<<0) != 0
		b := v&(1<<1) != 0
		switch {
		case a && b:
			i++
		}
	}

	return i
}
