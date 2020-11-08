package hyperspace

import (
	"sort"

	"github.com/twmb/murmur3"
)

type (
	Bloom []uint64
)

const (
	noOfHashes = 3
)

func New(content ...string) Bloom {
	bs := []uint64{}
	for _, c := range content {
		b := hash([]byte(c))
		bs = append(bs, b...)
	}
	bs = unique(bs)
	sort.Sort(uint64Sort(bs))
	return bs
}

func unique(s []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(s))
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

func (b Bloom) Test(q Bloom) bool {
	return intersectionCount(q, b) == len(q)
}

func intersectionCount(a, b []uint64) int {
	m := make(map[uint64]uint64)
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
		if a && b {
			i++
		}
	}

	return i
}

func hash(b []byte) []uint64 {
	h := make([]uint64, noOfHashes)
	for i := range h {
		h[i] = uint64(murmur3.SeedSum32(uint32(i), b))
	}
	return h
}

type uint64Sort []uint64

func (a uint64Sort) Len() int           { return len(a) }
func (a uint64Sort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a uint64Sort) Less(i, j int) bool { return a[i] < a[j] }
