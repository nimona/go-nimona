package object

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"sort"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"

	"nimona.io/pkg/errors"
)

const (
	cidCodec = 0x6e6d // nm codec
)

func mhFromBytes(t Hint, d []byte) (multihash.Multihash, error) {
	b := sha256.Sum256(append([]byte(t), d...))
	h, err := multihash.Encode(b[:], multihash.SHA2_256)
	if err != nil {
		panic(err)
	}
	return h, nil
}

// a v1 cid consists of:
// - <multibase-prefix>
// - <cid-version>
// - <multicodec-content-type>
// - <multihash-content>
func mhToCid(h multihash.Multihash) CID {
	c := cid.NewCidV1(cidCodec, h)
	// nolint: errcheck
	// there is nothing that can go wrong here
	s, _ := multibase.Encode(multibase.Base32, c.Bytes())
	return CID(s)
}

func mhFromCid(h CID) (multihash.Multihash, error) {
	c, err := cid.Decode(string(h))
	if err != nil {
		return nil, err
	}
	if c.Prefix().Codec != cidCodec {
		return nil, errors.Error("invalid cid codec")
	}
	return c.Hash(), nil
}

func NewCID(o *Object) (CID, error) {
	r, err := fromValue(o)
	if err != nil || r == nil {
		return EmptyCID, err
	}
	return mhToCid(r), nil
}

func fromValue(v Value) (multihash.Multihash, error) {
	switch vv := v.(type) {
	case Bool:
		if !vv {
			return mhFromBytes(BoolHint, []byte{0})
		}
		return mhFromBytes(BoolHint, []byte{1})
	case Data:
		return mhFromBytes(DataHint, vv)
	case Float:
		// replacing ben's implementation with something less custom, based on:
		// * https://github.com/benlaurie/objecthash
		// * https://play.golang.org/p/3xraud43pi
		// examples of same results in other languages
		// * ruby: `[7.30363941192626953125].pack('G').unpack('B*').first`
		// * js: `http://weitz.de/ieee`
		//
		// NOTE(geoah): I have removed the inf and nan hashing for now,
		// we can revisit them once we better understand their usecases.
		switch {
		case math.IsInf(float64(vv), 1):
			return nil, errors.Error("float inf is not currently supported")
		case math.IsInf(float64(vv), -1):
			return nil, errors.Error("float -inf is not currently supported")
		case math.IsNaN(float64(vv)):
			return nil, errors.Error("float nan is not currently supported")
		default:
			return mhFromBytes(FloatHint,
				[]byte(
					fmt.Sprintf(
						"%d",
						math.Float64bits(float64(vv)),
					),
				),
			)
		}
	case Int:
		return mhFromBytes(
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
			f = CIDHint
		} else {
			m = v.(Map)
			f = MapHint
		}
		h := multihash.Multihash{}
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
				mkf = CIDHint
			}
			vh, err := fromValue(mk)
			if err != nil {
				return nil, err
			}
			if vh == nil {
				continue
			}

			k = k + ":" + string(mkf)
			kh, err := mhFromBytes(
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
		return mhFromBytes(
			f,
			h,
		)
	case String:
		return mhFromBytes(
			StringHint,
			[]byte(string(vv)),
		)
	case Uint:
		return mhFromBytes(
			UintHint,
			[]byte(
				fmt.Sprintf(
					"%d",
					uint64(vv),
				),
			),
		)
	case CID:
		return mhFromCid(vv)
	case BoolArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(BoolArrayHint, h)
	case DataArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(DataArrayHint, h)
	case FloatArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(FloatArrayHint, h)
	case IntArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(IntArrayHint, h)
	case MapArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(MapArrayHint, h)
	case ObjectArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(ObjectArrayHint, h)
	case StringArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(StringArrayHint, h)
	case UintArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := fromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(UintArrayHint, h)
	case CIDArray:
		h := multihash.Multihash{}
		for _, ivv := range vv {
			ivvh, err := mhFromCid(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, ivvh...)
		}
		return mhFromBytes(CIDArrayHint, h)
	}
	panic("unknown value " + reflect.TypeOf(v).Name())
}
