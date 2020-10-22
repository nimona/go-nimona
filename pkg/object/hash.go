package object

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"nimona.io/internal/encoding/base58"
	"nimona.io/pkg/errors"
)

// ObjectHash consistently hashes a map.
// It is based on Ben Laurie's object hash, but using the same type hints
// as TJSON instead.

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

func NewHash(o *Object) (Hash, error) {
	m, err := objectToMap(o)
	if err != nil {
		return "", err
	}
	return hashMap(m)
}

func hintsFromKey(k string) []TypeHint {
	ps := strings.Split(k, ":")
	if len(ps) == 1 {
		return nil
	}
	hs := []TypeHint{}
	for _, sh := range ps[1] {
		hs = append(hs, GetTypeHint(string(sh)))
	}
	return hs
}

func hintFromKey(k string) string {
	ps := strings.Split(k, ":")
	if len(ps) == 1 {
		return ""
	}
	return ps[1]
}

func hashMap(m map[string]interface{}) (Hash, error) {
	if len(m) == 0 {
		return EmptyHash, nil
	}
	b := []byte{}
	ks := []string{}
	for k := range m {
		if len(k) > 0 && k[0] == '_' {
			continue
		}
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		v := m[k]
		if v == nil {
			continue
		}
		ts := hintsFromKey(k)
		hv, err := hashValueAs(k, v, ts...)
		if err != nil {
			return EmptyHash, err
		}
		if hv.IsEmpty() {
			continue
		}
		// hash the key (including the hint)
		// but if we are dealing with an object, replace its hint with r
		ck := k
		if strings.HasSuffix(k, ":"+HintObject.String()) {
			ck = k[:len(k)-2] + ":" + HintRef.String()
		}
		if strings.HasSuffix(k, ":"+HintMap.String()) {
			ck = k[:len(k)-2] + ":" + HintRef.String()
		}
		hk, _ := hash("", []byte(ck)) // nolint: errcheck
		b = append(b, hk...)
		b = append(b, hv...)
	}
	if len(b) == 0 {
		return EmptyHash, nil
	}
	return hash(HintRef, b)
}

func hash(p TypeHint, v []byte) (Hash, error) {
	d := sha256.Sum256(append([]byte(p), v...))
	return Hash("oh1." + base58.Encode(d[:])), nil
}

func hashValueAs(k string, o interface{}, ts ...TypeHint) (Hash, error) {
	if o == nil {
		return EmptyHash, nil
	}

	v := reflect.ValueOf(o)
	t := reflect.TypeOf(o)

	if len(ts) == 0 {
		return EmptyHash, nil
	}

	switch ts[0] {
	case HintArray:
		if v.Len() == 0 {
			return EmptyHash, nil
		}
		vs := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			vs = append(vs, v.Index(i).Interface())
		}
		b := []byte{}
		for _, iv := range vs {
			ivv, err := hashValueAs(k, iv, ts[1:]...)
			if err != nil {
				return EmptyHash, err
			}
			b = append(b, ivv...)
		}
		// TODO(geoah) hint SHOULD NOT be array, but array + inner hint
		return hash(TypeHint(hintFromKey(k)), b)
	case HintString:
		s, ok := o.(string)
		if !ok {
			if ss, ok := o.(interface{ String() string }); ok {
				s = ss.String()
			} else {
				return EmptyHash, nil
			}
		}
		if s == "" {
			return EmptyHash, nil
		}
		return hash(HintString, []byte(s))
	case HintData:
		switch t.Kind() {
		case reflect.String:
			d, err := base64.StdEncoding.DecodeString(o.(string))
			if err != nil {
				panic(err)
			}
			return hash(HintData, d)
		case reflect.Slice:
			bo := make([]byte, v.Len())
			switch t.Elem().Kind() {
			case reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64:
				for i := 0; i < v.Len(); i++ {
					bo[i] = uint8(v.Index(i).Int())
				}
				return hash(HintData, bo)
			case reflect.Uint,
				reflect.Uint8,
				reflect.Uint16,
				reflect.Uint32,
				reflect.Uint64:
				for i := 0; i < v.Len(); i++ {
					bo[i] = uint8(v.Index(i).Uint())
				}
				return hash(HintData, bo)
			case reflect.Interface:
				for i := 0; i < v.Len(); i++ {
					iv := v.Index(i).Interface()
					switch ivv := iv.(type) {
					case uint8:
						bo[i] = ivv
					case uint64:
						bo[i] = uint8(ivv)
					case float64:
						bo[i] = uint8(ivv)
					default:
						panic("data should be some sort of number array, was " +
							t.Elem().Kind().String())
					}
				}
				return hash(HintData, bo)
			default:
				panic("data should be some sort of number array, was " +
					t.Elem().Kind().String())
			}
		}
		return hash(HintData, o.([]byte))
	case HintMap, HintObject:
		switch v := o.(type) {
		case map[string]interface{}:
			if len(v) == 0 {
				return EmptyHash, nil
			}
			h, err := hashMap(v)
			if err != nil {
				return EmptyHash, err
			}
			return h, nil
		case *Object:
			m := v.ToMap()
			if len(m) == 0 {
				return EmptyHash, nil
			}
			h, err := hashMap(m)
			if err != nil {
				return EmptyHash, err
			}
			return h, nil
		default:
			panic("hashing only supports map[string]interface{}")
		}
	case HintFloat:
		switch t.Kind() {
		case reflect.Float32,
			reflect.Float64:
			nf, err := hashFloat(v.Float())
			if err != nil {
				return EmptyHash, err
			}
			return nf, nil
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			nf, err := hashFloat(float64(v.Int()))
			if err != nil {
				return EmptyHash, err
			}
			return nf, nil
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			nf, err := hashFloat(float64(v.Uint()))
			if err != nil {
				return EmptyHash, err
			}
			return nf, nil
		}
	case HintInt:
		switch t.Kind() {
		case reflect.Float32,
			reflect.Float64:
			return hash(HintInt, []byte(fmt.Sprintf("%d", int64(v.Float()))))
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			return hash(HintInt, []byte(fmt.Sprintf("%d", v.Int())))
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			return hash(HintInt, []byte(fmt.Sprintf("%d", int64(v.Uint()))))
		}
	case HintUint:
		return hash(HintUint, []byte(fmt.Sprintf("%d", o)))
	case HintBool:
		if v.Bool() {
			return hash(HintBool, []byte{1})
		}
		return hash(HintBool, []byte{0})
	case HintRef:
		switch h := o.(type) {
		case string:
			return Hash(h), nil
		case Hash:
			return h, nil
		}
		return EmptyHash, errors.Error("unsupported type for ref")
	}
	panic(
		fmt.Sprintf("hash: unsupported type %s (%s) for key %s",
			string(ts[0]),
			t.Kind().String(),
			k,
		),
	)
}

// replacing ben's implementation with something less custom, based on:
// * https://github.com/benlaurie/objecthash
// * https://play.golang.org/p/3xraud43pi
// examples of same results in other languages
// * ruby: `[7.30363941192626953125].pack('G').unpack('B*').first`
// * js: `http://weitz.de/ieee`
func hashFloat(f float64) (Hash, error) {
	nf := ""
	switch {
	case math.IsInf(f, 1):
		nf = "Infinity"
	case math.IsInf(f, -1):
		nf = "-Infinity"
	case math.IsNaN(f):
		nf = "NaN"
	default:
		nf = fmt.Sprintf("%x", math.Float64bits(f))
	}

	return hash(HintFloat, []byte(nf))
}
