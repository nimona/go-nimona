package object

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"sort"

	"nimona.io/internal/encoding/base58"
)

type (
	rawHash []byte
)

func fromBytes(t Hint, b []byte) (rawHash, error) {
	h := sha256.Sum256(append([]byte(t), b...))
	return h[:], nil
}

func hashFromRaw(r rawHash) Hash {
	return Hash("oh1." + base58.Encode(r))
}

func hashToRaw(h Hash) (rawHash, error) {
	b, err := base58.Decode(string(h)[4:])
	return rawHash(b), err
}

func NewHash(o *Object) (Hash, error) {
	r, err := fromValue(o)
	if err != nil {
		return EmptyHash, err
	}
	return hashFromRaw(r), nil
}

func fromValue(v Value) (rawHash, error) {
	switch vv := v.(type) {
	case Bool:
		if !vv {
			return fromBytes(BoolHint, []byte{0})
		}
		return fromBytes(BoolHint, []byte{1})
	case Data:
		return fromBytes(DataHint, vv)
	case Float:
		// replacing ben's implementation with something less custom, based on:
		// * https://github.com/benlaurie/objecthash
		// * https://play.golang.org/p/3xraud43pi
		// examples of same results in other languages
		// * ruby: `[7.30363941192626953125].pack('G').unpack('B*').first`
		// * js: `http://weitz.de/ieee`
		switch {
		case math.IsInf(float64(vv), 1):
			return fromBytes(FloatHint, []byte("Infinity"))
		case math.IsInf(float64(vv), -1):
			return fromBytes(FloatHint, []byte("-Infinity"))
		case math.IsNaN(float64(vv)):
			return fromBytes(FloatHint, []byte("NaN"))
		default:
			return fromBytes(FloatHint,
				[]byte(
					fmt.Sprintf(
						"%d",
						math.Float64bits(float64(vv)),
					),
				),
			)
		}
	case Int:
		return fromBytes(
			IntHint,
			[]byte(
				fmt.Sprintf(
					"%d",
					int64(vv),
				),
			),
		)
	case *Object, Map:
		var f Hint
		var m Map
		if o, ok := vv.(*Object); ok {
			m = o.Map()
			f = HashHint
		} else {
			m = v.(Map)
			f = MapHint
		}
		h := rawHash{}
		ks := []string{}
		for k := range m {
			if len(k) > 0 && k[0] == '_' {
				continue
			}
			ks = append(ks, k)
		}
		sort.Strings(ks)
		for _, k := range ks {
			mk := m[k]
			if mk == nil {
				continue
			}
			mkf := mk.Hint()
			if _, ok := mk.(*Object); ok {
				mkf = HashHint
			}
			vh, err := fromValue(mk)
			if err != nil {
				return nil, err
			}
			if vh == nil {
				continue
			}

			k = k + ":" + string(mkf)
			kh, err := fromBytes(
				StringHint,
				[]byte(k),
			)
			if err != nil {
				return nil, err
			}
			h = append(
				h,
				kh...,
			)
			h = append(
				h,
				vh...,
			)
		}
		if len(h) == 0 {
			return nil, nil
		}
		return fromBytes(
			f,
			h,
		)
	case String:
		return fromBytes(
			StringHint,
			[]byte(string(vv)),
		)
	case Uint:
		return fromBytes(
			UintHint,
			[]byte(
				fmt.Sprintf(
					"%d",
					uint64(vv),
				),
			),
		)
	case Hash:
		return hashToRaw(vv)
	case BoolArray:
		h := rawHash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return fromBytes(BoolArrayHint, h)
	case DataArray:
		h := rawHash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return fromBytes(DataArrayHint, h)
	case FloatArray:
		h := rawHash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return fromBytes(FloatArrayHint, h)
	case IntArray:
		h := rawHash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return fromBytes(IntArrayHint, h)
	case MapArray:
		h := rawHash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return fromBytes(MapArrayHint, h)
	case ObjectArray:
		h := rawHash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return fromBytes(ObjectArrayHint, h)
	case StringArray:
		h := rawHash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return fromBytes(StringArrayHint, h)
	case UintArray:
		h := rawHash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return fromBytes(UintArrayHint, h)
	}
	panic("unknown value " + reflect.TypeOf(v).Name())
}
