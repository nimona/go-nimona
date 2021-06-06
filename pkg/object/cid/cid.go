package cid

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
	"nimona.io/pkg/object/hint"
	"nimona.io/pkg/object/value"
)

const (
	cidCodec = 0x6E6D // codec code for nimona object
)

func New(o value.Value) (value.CID, error) {
	r, err := FromValue(o)
	if err != nil {
		return Invalid, err
	}
	if r == nil {
		return Empty, nil
	}
	return mhToCid(r), nil
}

func Must(v value.CID, err error) value.CID {
	if err != nil {
		panic(err)
	}
	return v
}

func mhFromBytes(t hint.Hint, d []byte) (multihash.Multihash, error) {
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
func mhToCid(h multihash.Multihash) value.CID {
	c := cid.NewCidV1(cidCodec, h)
	// nolint: errcheck
	// there is nothing that can go wrong here
	s, _ := multibase.Encode(multibase.Base32, c.Bytes())
	return value.CID(s)
}

func mhFromCid(h value.CID) (multihash.Multihash, error) {
	c, err := cid.Decode(string(h))
	if err != nil {
		return nil, err
	}
	if c.Prefix().Codec != cidCodec {
		return nil, errors.Error("invalid cid codec")
	}
	return c.Hash(), nil
}

func FromValue(v value.Value) (multihash.Multihash, error) {
	switch vv := v.(type) {
	case value.Bool:
		if !vv {
			return mhFromBytes(hint.Bool, []byte{0})
		}
		return mhFromBytes(hint.Bool, []byte{1})
	case value.Data:
		return mhFromBytes(hint.Data, vv)
	case value.Float:
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
			return mhFromBytes(hint.Float,
				[]byte(
					fmt.Sprintf(
						"%d",
						math.Float64bits(float64(vv)),
					),
				),
			)
		}
	case value.Int:
		return mhFromBytes(
			hint.Int,
			[]byte(
				fmt.Sprintf(
					"%d",
					int64(vv),
				),
			),
		)
	case value.Map:
		m := v.(value.Map)
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
			mkf := m.Hint()
			if _, ok := mk.(value.Map); ok {
				mkf = hint.Map
			}
			vh, err := FromValue(mk)
			if err != nil {
				return nil, err
			}
			if vh == nil {
				continue
			}

			k = k + ":" + string(mkf)
			kh, err := mhFromBytes(
				hint.String,
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
			hint.CID, // TODO(geoah) should this be hint.Map?
			h,
		)
	case value.String:
		if string(vv) == "" {
			return nil, nil
		}
		return mhFromBytes(
			hint.String,
			[]byte(string(vv)),
		)
	case value.Uint:
		return mhFromBytes(
			hint.Uint,
			[]byte(
				fmt.Sprintf(
					"%d",
					uint64(vv),
				),
			),
		)
	case value.CID:
		if vv == "" {
			return nil, nil
		}
		return mhFromCid(vv)
	case value.BoolArray:
		if vv.Len() == 0 {
			return nil, nil
		}
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := FromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(hint.BoolArray, h)
	case value.DataArray:
		if vv.Len() == 0 {
			return nil, nil
		}
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := FromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(hint.DataArray, h)
	case value.FloatArray:
		if vv.Len() == 0 {
			return nil, nil
		}
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := FromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(hint.FloatArray, h)
	case value.IntArray:
		if vv.Len() == 0 {
			return nil, nil
		}
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := FromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(hint.IntArray, h)
	case value.MapArray:
		if vv.Len() == 0 {
			return nil, nil
		}
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := FromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(hint.MapArray, h)
	case value.StringArray:
		if vv.Len() == 0 {
			return nil, nil
		}
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := FromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(hint.StringArray, h)
	case value.UintArray:
		if vv.Len() == 0 {
			return nil, nil
		}
		h := multihash.Multihash{}
		for _, ivv := range vv {
			vh, err := FromValue(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, vh...)
		}
		return mhFromBytes(hint.UintArray, h)
	case value.CIDArray:
		if vv.Len() == 0 {
			return nil, nil
		}
		h := multihash.Multihash{}
		for _, ivv := range vv {
			ivvh, err := mhFromCid(ivv)
			if err != nil {
				return nil, err
			}
			h = append(h, ivvh...)
		}
		return mhFromBytes(hint.CIDArray, h)
	}
	panic("unknown value " + reflect.TypeOf(v).Name())
}
