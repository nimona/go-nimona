package object

import "sort"

const (
	EmptyHash = Hash("")
)

// SortHashes sorts a slice of Hashes in increasing order, and also returns it.
// The return part is mostly for allowing this to be used as a helper method in
// tests.
func SortHashes(a []Hash) []Hash {
	sort.Sort(HashSlice(a))
	return a
}

type HashSlice []Hash

func (p HashSlice) Len() int           { return len(p) }
func (p HashSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p HashSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p HashSlice) Sort()              { sort.Sort(p) }

func (h Hash) IsEmpty() bool {
	return string(h) == ""
}
