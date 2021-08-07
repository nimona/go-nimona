package tilde

import (
	"crypto/sha256"
	"sort"

	"nimona.io/internal/encoding/base58"
)

const (
	EmptyDigest Digest = ""
)

func (v Digest) Hint() Hint {
	return DigestHint
}

func (v Digest) _isValue() {
}

func (v Digest) Hash() Digest {
	return v
}

func hashFromBytes(d []byte) Digest {
	if d == nil {
		return EmptyDigest
	}
	b := sha256.Sum256(d)
	return Digest(base58.Encode(b[:]))
}

func (v Digest) Bytes() ([]byte, error) {
	return base58.Decode(string(v))
}

func (v Digest) IsEmpty() bool {
	return string(v) == ""
}

func (v Digest) Equal(h Digest) bool {
	return h == v
}

func (v Digest) String() string {
	return string(v)
}

// SortDigests sorts a slice of Digestes in increasing order, and returns it.
// The return part is mostly for allowing this to be used as a helper method in
// tests.
func SortDigests(a []Digest) []Digest {
	sort.Sort(Slice(a))
	return a
}

type Slice []Digest

func (p Slice) Len() int           { return len(p) }
func (p Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Slice) Sort()              { sort.Sort(p) }
