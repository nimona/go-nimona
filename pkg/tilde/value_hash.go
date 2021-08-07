package tilde

import (
	"crypto/sha256"
	"sort"

	"nimona.io/internal/encoding/base58"
)

const (
	EmptyHash Hash = ""
)

func (v Hash) Hint() Hint {
	return HashHint
}

func (v Hash) _isValue() {
}

func (v Hash) Hash() Hash {
	return v
}

func hashFromBytes(d []byte) Hash {
	if d == nil {
		return EmptyHash
	}
	b := sha256.Sum256(d)
	return Hash(base58.Encode(b[:]))
}

func (v Hash) Bytes() ([]byte, error) {
	return base58.Decode(string(v))
}

func (v Hash) IsEmpty() bool {
	return string(v) == ""
}

func (v Hash) Equal(h Hash) bool {
	return h == v
}

func (v Hash) String() string {
	return string(v)
}

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
