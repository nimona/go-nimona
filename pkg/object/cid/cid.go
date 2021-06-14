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

	"nimona.io/pkg/chore"
	"nimona.io/pkg/errors"
)

const (
	cidCodec = 0x6E6D // codec code for nimona object
)

func New(o chore.Value) (chore.CID, error) {
	r, err := FromValue(o)
	if err != nil {
		return Invalid, err
	}
	if r == nil {
		return Empty, nil
	}
	return mhToCid(r), nil
}

func Must(v chore.CID, err error) chore.CID {
	if err != nil {
		panic(err)
	}
	return v
}

func mhFromBytes(t chore.Hint, d []byte) (multihash.Multihash, error) {
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
func mhToCid(h multihash.Multihash) chore.CID {
	c := cid.NewCidV1(cidCodec, h)
	// nolint: errcheck
	// there is nothing that can go wrong here
	s, _ := multibase.Encode(multibase.Base32, c.Bytes())
	return chore.CID(s)
}

func mhFromCid(h chore.CID) (multihash.Multihash, error) {
	c, err := cid.Decode(string(h))
	if err != nil {
		return nil, err
	}
	if c.Prefix().Codec != cidCodec {
		return nil, errors.Error("invalid cid codec")
	}
	return c.Hash(), nil
}

func FromValue(v chore.Value) (multihash.Multihash, error) {
	switch vv := v.(type) {
	case chore.Bool:
		if !vv {
			return mhFromBytes(chore.BoolHint, []byte{0})
		}
		return mhFromBytes(chore.BoolHint, []byte{1})
	case chore.Data:
		return mhFromBytes(chore.DataHint, vv)
	case chore.Float:
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
			return mhFromBytes(chore.FloatHint,
				[]byte(
					fmt.Sprintf(
						"%d",
						math.Float64bits(float64(vv)),
					),
				),
			)
		}
	case chore.Int:
		return mhFromBytes(
			chore.IntHint,
			[]byte(
				fmt.Sprintf(
					"%d",
					int64(vv),
				),
			),
		)
	case chore.Map:
		m := v.(chore.Map)
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
			if _, ok := mk.(chore.Map); ok {
				mkf = chore.MapHint
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
				chore.StringHint,
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
			chore.CIDHint, // TODO(geoah) should this be chore.MapHint?
			h,
		)
	case chore.String:
		if string(vv) == "" {
			return nil, nil
		}
		return mhFromBytes(
			chore.StringHint,
			[]byte(string(vv)),
		)
	case chore.Uint:
		return mhFromBytes(
			chore.UintHint,
			[]byte(
				fmt.Sprintf(
					"%d",
					uint64(vv),
				),
			),
		)
	case chore.CID:
		if vv == "" {
			return nil, nil
		}
		return mhFromCid(vv)
	case chore.BoolArray:
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
		return mhFromBytes(chore.BoolArrayHint, h)
	case chore.DataArray:
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
		return mhFromBytes(chore.DataArrayHint, h)
	case chore.FloatArray:
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
		return mhFromBytes(chore.FloatArrayHint, h)
	case chore.IntArray:
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
		return mhFromBytes(chore.IntArrayHint, h)
	case chore.MapArray:
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
		return mhFromBytes(chore.MapArrayHint, h)
	case chore.StringArray:
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
		return mhFromBytes(chore.StringArrayHint, h)
	case chore.UintArray:
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
		return mhFromBytes(chore.UintArrayHint, h)
	case chore.CIDArray:
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
		return mhFromBytes(chore.CIDArrayHint, h)
	}
	panic("unknown value " + reflect.TypeOf(v).Name())
}
