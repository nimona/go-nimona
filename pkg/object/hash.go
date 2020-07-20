package object

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sort"
	"strings"

	"nimona.io/internal/encoding/base58"
)

type (
	Hash string
)

func (h Hash) IsEmpty() bool {
	return h == ""
}

func (h Hash) IsEqual(c Hash) bool {
	return h == c
}

func (h Hash) String() string {
	return string(h)
}

func (h Hash) rawBytes() []byte {
	ps := strings.Split(string(h), ".")
	if len(ps) == 0 {
		return nil
	}
	p := ps[len(ps)-1]
	b, _ := base58.Decode(p)
	return b
}

func hash(p TypeHint, v []byte) Hash {
	d := sha256.Sum256(append([]byte(p), v...))
	return Hash("oh1." + base58.Encode(d[:]))
}

func (v Bool) Hash() Hash   { return v.hash() }
func (v String) Hash() Hash { return v.hash() }
func (v Int) Hash() Hash    { return v.hash() }
func (v Float) Hash() Hash  { return v.hash() }
func (v Bytes) Hash() Hash  { return v.hash() }
func (v Map) Hash() Hash    { return v.hash() }
func (v List) Hash() Hash   { return v.hash() }

// refs don't get re-hashed
func (v Ref) Hash() Hash { return Hash(v) }

func (v Bool) hash() Hash {
	if v {
		return hash("b", []byte{1})
	}
	return hash("b", []byte{0})
}

func (v String) hash() Hash {
	return hash("s", []byte(v))
}

func (v Int) hash() Hash {
	return hash("i",
		[]byte(fmt.Sprintf("%d", int64(v))))
}

func (v Float) hash() Hash {
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

func (v Bytes) hash() Hash {
	return hash("d", v)
}

// TODO for now we treat maps and objects the same, so we convert their hints
// into `r`s; Once we split them, map hints should be kept as `m`.
func (v Map) hash() Hash {
	// get all map keys
	ks := []string{}
	v.Iterate(func(k string, v Value) bool {
		if !strings.HasPrefix(k, "_") {
			ks = append(ks, k)
		}
		return true
	})
	// return if there are no keys we can use
	if len(ks) == 0 {
		return ""
	}
	// sort them
	sort.Strings(ks)
	h := []byte{}
	// go through all keys and values and add their hashes together
	for _, k := range ks {
		// hash the key (including the hint)
		// but if we are dealing with an object, replace its hint
		ck := k
		if strings.HasSuffix(k, ":"+HintMap.String()) {
			ck = k[:len(k)-2] + ":" + HintRef.String()
		}
		// hash the value
		vh := v.Value(k).Hash()
		// and move on if nothing was hashed
		if vh.IsEmpty() {
			continue
		}
		// else, append the key's hash
		h = append(h, hash("", []byte(ck))...)
		// and finally append the value hash
		h = append(h, vh...)
	}
	// and finally hash the whole thing
	return hash("r", h)
}

func (v List) hash() Hash {
	h := []byte{}
	v.Iterate(func(v Value) bool {
		h = append(h, v.Hash()...)
		return true
	})
	return hash(v.hint, h)
}
