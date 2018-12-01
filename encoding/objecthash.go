package encoding

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"sort"
)

// ObjectHash consistently hashes a map.
// It is based on Ben Laurie's object hash, but using the same type hints
// as TJSON instead.
// TODO add redaction
func ObjectHash(o *Object) ([]byte, error) {
	m := o.Map()
	b := []byte{}
	ks := []string{}
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	x := map[string]interface{}{}
	for _, k := range ks {
		v := m[k]
		hk := hash(HintString, []byte(k))
		hv := hashValue(v)
		b = append(b, hk...)
		b = append(b, hv...)
		x[k] = hv
	}
	return hash(HintMap, b), nil
}

func hash(p string, b []byte) []byte {
	h := sha256.New()
	h.Write([]byte(p))
	h.Write(b)
	return h.Sum(nil)
}

func hashValue(o interface{}) []byte {
	v := reflect.ValueOf(o)
	t := reflect.TypeOf(o)
	switch t.Kind() {
	case reflect.Invalid: // nil
		return hash(HintNil, []byte{})
	case reflect.Slice, reflect.Array:
		if t.Elem() == reflect.TypeOf(byte(0)) {
			return hash(HintBytes, o.([]byte))
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
		return hash(HintString, []byte(o.(string)))
	case reflect.Struct:
		panic("structs are not currently supported")
	case reflect.Map:
		m, ok := o.(map[string]interface{})
		if !ok {
			panic("hashing only supports map[string]interface{}")
		}
		o := NewObjectFromMap(m)
		h, err := ObjectHash(o)
		if err != nil {
			panic("hashing error: " + err.Error())
		}
		fmt.Printf("nested hash: %x\n", h)
		return h
	case reflect.Float32, reflect.Float64:
		nf, err := floatNormalize(v.Float())
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

// from https://github.com/benlaurie/objecthash/blob/c7d617cadf0865f370d65cc82796c2e05506d26a/go/objecthash/objecthash.go#L115
func floatNormalize(originalFloat float64) (s string, err error) {
	// s = fmt.Sprintf("%b", originalFloat)
	// fmt.Println(">>>>>", originalFloat, s)
	// return

	// Special case 0
	// Note that if we allowed f to end up > .5 or == 0, we'd get the same thing.
	if originalFloat == 0 {
		return "+0:", nil
	}

	// sign
	f := originalFloat
	s = `+`
	if f < 0 {
		s = `-`
		f = -f
	}

	// exponent
	e := 0
	for f > 1 {
		f /= 2
		e++
	}
	for f <= .5 {
		f *= 2
		e--
	}
	s += fmt.Sprintf("%d:", e)

	// mantissa
	if f > 1 || f <= .5 {
		return "", fmt.Errorf("Could not normalize float: %f", originalFloat)
	}
	for f != 0 {
		if f >= 1 {
			s += `1`
			f--
		} else {
			s += `0`
		}
		if f >= 1 {
			return "", fmt.Errorf("Could not normalize float: %f", originalFloat)
		}
		if len(s) >= 1000 {
			return "", fmt.Errorf("Could not normalize float: %f", originalFloat)
		}
		f *= 2
	}
	fmt.Println(">>>>>", originalFloat, s)
	return
}

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
		var err error
		nf, err = floatNormalize(f)
		if err != nil {
			return nil, err
		}
	}

	return hash(HintFloat, []byte(nf)), nil
}
