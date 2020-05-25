package immutable

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sort"
	"strings"

	"nimona.io/internal/encoding/base58"
)

func hash(p string, v []byte) []byte {
	d := sha256.Sum256(append([]byte(p), v...))
	return d[:]
}

func formatHash(b []byte) string {
	return "oh1." + base58.Encode(b)
}

func (v Bool) Hash() string   { return formatHash(v.hash()) }
func (v String) Hash() string { return formatHash(v.hash()) }
func (v Int) Hash() string    { return formatHash(v.hash()) }
func (v Float) Hash() string  { return formatHash(v.hash()) }
func (v Bytes) Hash() string  { return formatHash(v.hash()) }
func (v Map) Hash() string    { return formatHash(v.hash()) }
func (v List) Hash() string   { return formatHash(v.hash()) }

func (v Bool) hash() []byte {
	if v {
		return hash("b", []byte{1})
	}
	return hash("b", []byte{0})
}

func (v String) hash() []byte {
	return hash("s", []byte(v))
}

func (v Int) hash() []byte {
	return hash("i",
		[]byte(fmt.Sprintf("%d", int64(v))))
}

func (v Float) hash() []byte {
	nf := ""
	switch {
	case math.IsInf(float64(v), 1):
		nf = "Infinity"
	case math.IsInf(float64(v), -1):
		nf = "-Infinity"
	case math.IsNaN(float64(v)):
		// TODO should this be even supported?
		nf = "NaN"
	default:
		nf = fmt.Sprintf("%x", math.Float64bits(float64(v)))
	}
	return hash("f", []byte(nf))
}

func (v Bytes) hash() []byte {
	return hash("d", v)
}

func (v Map) hash() []byte {
	// get all map keys
	ks := []string{}
	v.Iterate(func(k string, v Value) {
		if strings.HasPrefix(k, "_") {
			return
		}
		ks = append(ks, k)
	})
	// sort them
	sort.Strings(ks)
	h := []byte{}
	// go through all keys and values and add their hashes together
	for _, k := range ks {
		// hash the key (including the hint)
		h = append(h, hash("", []byte(k))...)
		// hash the value
		h = append(h, v.Value(k).hash()...)
	}
	// and finally hash the whole thing
	return hash("o", h)
}

func (v List) hash() []byte {
	h := []byte{}
	v.Iterate(func(v Value) {
		h = append(h, v.hash()...)
	})
	return hash(v.hint, h)
}
