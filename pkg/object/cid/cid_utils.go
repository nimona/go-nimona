package cid

import (
	"sort"

	"nimona.io/pkg/chore"
)

const (
	Empty   = chore.CID("")
	Invalid = chore.CID("invalid")
)

// SortCIDs sorts a slice of CIDs in increasing order, and also returns it.
// The return part is mostly for allowing this to be used as a helper method in
// tests.
func SortCIDs(a []chore.CID) []chore.CID {
	sort.Sort(Slice(a))
	return a
}

type Slice []chore.CID

func (p Slice) Len() int           { return len(p) }
func (p Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Slice) Sort()              { sort.Sort(p) }
