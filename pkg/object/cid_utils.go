package object

import "sort"

const (
	EmptyCID = CID("")
)

// SortCIDs sorts a slice of CIDs in increasing order, and also returns it.
// The return part is mostly for allowing this to be used as a helper method in
// tests.
func SortCIDs(a []CID) []CID {
	sort.Sort(CIDSlice(a))
	return a
}

type CIDSlice []CID

func (p CIDSlice) Len() int           { return len(p) }
func (p CIDSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p CIDSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p CIDSlice) Sort()              { sort.Sort(p) }

func (h CID) IsEmpty() bool {
	return string(h) == ""
}
