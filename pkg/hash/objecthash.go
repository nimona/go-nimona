package hash

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"nimona.io/internal/encoding/base58"
	"nimona.io/pkg/object"
)

type (
	contentHash struct {
		algorithm string `json:"algorithm:s,omitempty"`
		d         []byte `json:"d:d,omitempty"`
	}
)

func FromBytes(b []byte) object.Hash {
	return formatHash(
		contentHash{
			algorithm: "oh1",
			d:         b,
		},
	)
}

func formatHash(h contentHash) object.Hash {
	s := h.algorithm + "." + base58.Encode(h.d)
	return object.Hash(s)
}

// New consistently hashes a map.
// It is based on Ben Laurie's object hash, but using the same type hints
// as TJSON instead.
// TODO add redaction
func New(o object.Object) object.Hash {
	d, err := objecthash(o.ToMap())
	if err != nil {
		panic(err)
	}
	// TODO(geoah) consider having an invalid hash type
	return formatHash(
		contentHash{
			algorithm: "oh1",
			d:         d,
		},
	)
}

func hintsFromKey(k string) []object.TypeHint {
	ps := strings.Split(k, ":")
	if len(ps) == 1 {
		return nil
	}
	hs := []object.TypeHint{}
	for _, sh := range ps[1] {
		hs = append(hs, object.GetTypeHint(string(sh)))
	}
	return hs
}

func objecthash(m map[string]interface{}) ([]byte, error) {
	b := []byte{}
	ks := []string{}
	for k := range m {
		if len(k) > 0 && k[0] == '_' {
			continue
		}
		ks = append(ks, k)
	}
	sort.Strings(ks)
	x := map[string]interface{}{}
	for _, k := range ks {
		v := m[k]
		if v == nil {
			continue
		}
		ts := hintsFromKey(k)
		hv := hashValueAs(k, v, ts...)
		if hv == nil {
			continue
		}
		// fmt.Println("hashing value for", k, "as", fmt.Sprintf("%x", hv))
		hk := hash(object.HintString, []byte(k))
		b = append(b, hk...)
		b = append(b, hv...)
		x[k] = hv
	}
	h := hash(object.HintObject, b)
	return h, nil
}

func hash(p object.TypeHint, b []byte) []byte {
	h := sha256.New()
	if _, err := h.Write([]byte(p)); err != nil {
		panic(err)
	}
	if _, err := h.Write(b); err != nil {
		panic(err)
	}
	return h.Sum(nil)
}

func hashValueAs(k string, o interface{}, ts ...object.TypeHint) []byte {
	if o == nil {
		return nil
	}

	v := reflect.ValueOf(o)
	t := reflect.TypeOf(o)

	if len(ts) == 0 {
		return nil
	}

	switch ts[0] {
	case object.HintArray:
		if v.Len() == 0 {
			return nil
		}
		vs := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			vs = append(vs, v.Index(i).Interface())
		}
		b := []byte{}
		for _, iv := range vs {
			b = append(b, hashValueAs(k, iv, ts[1:]...)...)
		}
		// TODO(geoah) hint SHOULD NOT be array, but array + inner hint
		return hash(object.HintArray, b)
	case object.HintString:
		s, ok := o.(string)
		if !ok {
			if ss, ok := o.(interface{ String() string }); ok {
				s = ss.String()
			} else {
				return nil
			}
		}
		if s == "" {
			return nil
		}
		return hash(object.HintString, []byte(s))
	case object.HintData:
		switch t.Kind() {
		case reflect.String:
			d, err := base64.StdEncoding.DecodeString(o.(string))
			if err != nil {
				panic(err)
			}
			return hash(object.HintData, d)
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
				return hash(object.HintData, bo)
			case reflect.Uint,
				reflect.Uint8,
				reflect.Uint16,
				reflect.Uint32,
				reflect.Uint64:
				for i := 0; i < v.Len(); i++ {
					bo[i] = uint8(v.Index(i).Uint())
				}
				return hash(object.HintData, bo)
			case reflect.Interface:
				for i := 0; i < v.Len(); i++ {
					iv := v.Index(i).Interface()
					switch iv.(type) {
					case uint8:
						bo[i] = iv.(uint8)
					case uint64:
						bo[i] = uint8(iv.(uint64))
					case float64:
						bo[i] = uint8(iv.(float64))
					default:
						panic("data should be some sort of number array, was " +
							t.Elem().Kind().String())
					}
				}
				return hash(object.HintData, bo)
			default:
				panic("data should be some sort of number array, was " +
					t.Elem().Kind().String())
			}
		}
		return hash(object.HintData, o.([]byte))
	case object.HintObject:
		m, ok := o.(map[string]interface{})
		if !ok {
			panic("hashing only supports map[string]interface{}")
		}
		if len(m) == 0 {
			return nil
		}
		h, err := objecthash(m)
		if err != nil {
			panic("hashing error: " + err.Error())
		}
		return h
	case object.HintFloat:
		switch t.Kind() {
		case reflect.Float32,
			reflect.Float64:
			nf, err := hashFloat(v.Float())
			if err != nil {
				panic(err)
			}
			return hash(object.HintFloat, nf)
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			nf, err := hashFloat(float64(v.Int()))
			if err != nil {
				panic(err)
			}
			return hash(object.HintFloat, nf)
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			nf, err := hashFloat(float64(v.Uint()))
			if err != nil {
				panic(err)
			}
			return hash(object.HintFloat, nf)
		}
	case object.HintInt:
		switch t.Kind() {
		case reflect.Float32,
			reflect.Float64:
			return hash(object.HintInt, []byte(fmt.Sprintf("%d", int64(v.Float()))))
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			return hash(object.HintInt, []byte(fmt.Sprintf("%d", v.Int())))
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			return hash(object.HintInt, []byte(fmt.Sprintf("%d", int64(v.Uint()))))
		}
	case object.HintUint:
		return hash(object.HintUint, []byte(fmt.Sprintf("%d", o)))
	case object.HintBool:
		if v.Bool() {
			return hash(object.HintBool, []byte{1})
		}
		return hash(object.HintBool, []byte{0})
	}
	panic(
		fmt.Sprintf("hash: unsupported type %s (%s) for key %s",
			string(ts[0]),
			t.Kind().String(),
			k,
		),
	)
}

// func hashValue(o interface{}) []byte {
// 	v := reflect.ValueOf(o)
// 	t := reflect.TypeOf(o)
// 	switch t.Kind() {
// 	case reflect.Invalid: // nil
// 		// return hash(HintNil, []byte{})
// 		return nil
// 	case reflect.Slice, reflect.Array:
// 		if v.Len() == 0 {
// 			return nil
// 		}
// 		if t.Elem() == reflect.TypeOf(byte(0)) {
// 			return hash(object.HintData, o.([]byte))
// 		}
// 		vs := []interface{}{}
// 		for i := 0; i < v.Len(); i++ {
// 			vs = append(vs, v.Index(i).Interface())
// 		}
// 		b := []byte{}
// 		for _, iv := range vs {
// 			b = append(b, hashValue(iv)...)
// 		}
// 		return hash(object.HintArray, b)
// 	case reflect.String:
// 		if o.(string) == "" {
// 			return nil
// 		}
// 		return hash(object.HintString, []byte(o.(string)))
// 	case reflect.Struct:
// 		panic("structs are not currently supported")
// 	case reflect.Map:
// 		m, ok := o.(map[string]interface{})
// 		if !ok {
// 			panic("hashing only supports map[string]interface{}")
// 		}
// 		if len(m) == 0 {
// 			return nil
// 		}
// 		h, err := objecthash(m, false)
// 		if err != nil {
// 			panic("hashing error: " + err.Error())
// 		}
// 		return h
// 	case reflect.Float32, reflect.Float64:
// 		nf, err := hashFloat(v.Float())
// 		if err != nil {
// 			panic(err)
// 		}
// 		return hash(object.HintFloat, []byte(nf))
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		return hash(object.HintInt, []byte(fmt.Sprintf("%d", v.Int())))
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		return hash(object.HintUint, []byte(fmt.Sprintf("%d", v.Uint())))
// 	case reflect.Bool:
// 		if v.Bool() {
// 			return hash(object.HintBool, []byte{1})
// 		}
// 		return hash(object.HintBool, []byte{0})
// 	}
// 	panic("hash: unsupported type " + v.String() + " -- " + fmt.Sprintf("%#v", o))
// }

// replacing ben's implementation with something less custom, based on:
// * https://github.com/benlaurie/objecthash
// * https://play.golang.org/p/3xraud43pi
// examples of same results in other languages
// * ruby: `[7.30363941192626953125].pack('G').unpack('B*').first`
// * js: `http://weitz.de/ieee`
func hashFloat(f float64) ([]byte, error) {
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

	return hash(object.HintFloat, []byte(nf)), nil
}
