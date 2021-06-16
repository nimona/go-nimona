package chore

import (
	"sort"
)

// SortHashes sorts a slice of Hashes in increasing order, and also returns it.
// The return part is mostly for allowing this to be used as a helper method in
// tests.
func SortHashes(a []Hash) []Hash {
	sort.Sort(Slice(a))
	return a
}

type Slice []Hash

func (p Slice) Len() int           { return len(p) }
func (p Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Slice) Sort()              { sort.Sort(p) }
