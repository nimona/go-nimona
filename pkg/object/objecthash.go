package object

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

// ObjectHash consistently hashes a map.
// It is based on Ben Laurie's object hash, but using the same type hints
// as TJSON instead.
// TODO add redaction
func ObjectHash(o *Object) ([]byte, error) {
	return objecthash(o, true)
}

func objecthash(o *Object, skipSig bool) ([]byte, error) {
	m := o.ToMap()
	b := []byte{}
	ks := []string{}
	for k := range m {
		// TODO(geoah) is there a better way of doing this?
		if k == "@" {
			continue
		}
		if skipSig && strings.HasPrefix(k, "@signature") {
			continue
		}
		ks = append(ks, k)
	}
	sort.Strings(ks)
	x := map[string]interface{}{}
	for _, k := range ks {
		v := m[k]
		hv := hashValue(v)
		if hv == nil {
			continue
		}
		hk := hash(HintString, []byte(k))
		b = append(b, hk...)
		b = append(b, hv...)
		x[k] = hv
	}
	h := hash(HintMap, b)
	return h, nil
}

func hash(p TypeHint, b []byte) []byte {
	h := sha256.New()
	if _, err := h.Write([]byte(p)); err != nil {
		panic(err)
	}
	if _, err := h.Write(b); err != nil {
		panic(err)
	}
	return h.Sum(nil)
}

func hashValue(o interface{}) []byte {
	v := reflect.ValueOf(o)
	t := reflect.TypeOf(o)
	switch t.Kind() {
	case reflect.Invalid: // nil
		// return hash(HintNil, []byte{})
		return nil
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil
		}
		if t.Elem() == reflect.TypeOf(byte(0)) {
			return hash(HintData, o.([]byte))
		}
		vs := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			vs = append(vs, v.Index(i).Interface())
		}
		b := []byte{}
		for _, iv := range vs {
			b = append(b, hashValue(iv)...)
		}
		return hash(HintArray, b)
	case reflect.String:
		if o.(string) == "" {
			return nil
		}
		return hash(HintString, []byte(o.(string)))
	case reflect.Struct:
		panic("structs are not currently supported")
	case reflect.Map:
		m, ok := o.(map[string]interface{})
		if !ok {
			panic("hashing only supports map[string]interface{}")
		}
		if len(m) == 0 {
			return nil
		}
		o := FromMap(m)
		h, err := objecthash(o, false)
		if err != nil {
			panic("hashing error: " + err.Error())
		}
		return h
	case reflect.Float32, reflect.Float64:
		nf, err := hashFloat(v.Float())
		if err != nil {
			panic(err)
		}
		return hash(HintFloat, []byte(nf))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return hash(HintInt, []byte(fmt.Sprintf("%d", v.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return hash(HintUint, []byte(fmt.Sprintf("%d", v.Uint())))
	case reflect.Bool:
		if v.Bool() {
			return hash(HintBool, []byte{1})
		}
		return hash(HintBool, []byte{0})
	}
	panic("hash: unsupported type " + v.String() + " -- " + fmt.Sprintf("%#v", o))
}

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

	return hash(HintFloat, []byte(nf)), nil
}
