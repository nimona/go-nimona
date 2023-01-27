package nimona

import (
	"bytes"
	"fmt"
	"strings"

	cbg "github.com/whyrusleeping/cbor-gen"
)

type DocumentMap map[string]interface{}

func NewDocumentMap(c Cborer) (m DocumentMap, err error) {
	b, err := MarshalCBORBytes(c)
	if err != nil {
		return m, fmt.Errorf("error marshaling cbor: %s", err)
	}

	return NewDocumentMapFromCBOR(b)
}

func NewDocumentMapFromCBOR(b []byte) (h DocumentMap, err error) {
	r := cbg.NewCborReader(bytes.NewReader(b))
	maj, n, err := r.ReadHeader()
	if err != nil {
		return h, fmt.Errorf("error reading header: %s", err)
	}

	if maj != cbg.MajMap {
		return h, fmt.Errorf("cannot hash non maps")
	}

	h, err = documentMap(r, n)
	if err != nil {
		return h, fmt.Errorf("error hashing map: %s", err)
	}

	return h, nil
}

// nolint: gocyclo // TODO: Refactor to reduce complexity
func documentMap(r *cbg.CborReader, extra uint64) (m DocumentMap, err error) {
	m = DocumentMap{}
	var v interface{}
	for i := uint64(0); i < extra; i++ {
		// read the key
		key, err := cbg.ReadString(r)
		if err != nil {
			return m, fmt.Errorf("error reading key: %w", err)
		}
		// skip ephemeral fields
		if strings.HasPrefix(key, "_") {
			continue
		}
		// read the value
		valMaj, extra, err := r.ReadHeader()
		if err != nil {
			return m, fmt.Errorf("error reading value header: %s", err)
		}
		switch valMaj {
		case cbg.MajMap:
			v, err = documentMap(r, extra)
		case cbg.MajUnsignedInt:
			v, err = MapUInt(r, extra)
		case cbg.MajNegativeInt:
			v, err = MapInt(r, extra)
		case cbg.MajByteString:
			if extra == 0 {
				continue
			}
			v, err = MapByteString(r, extra)
		case cbg.MajTextString:
			if extra == 0 {
				continue
			}
			v, err = MapTextString(r, extra)
		case cbg.MajArray:
			if extra == 0 {
				continue
			}
			v, err = MapArray(r, extra)
		case cbg.MajTag:
			panic("tags not supported")
		case cbg.MajOther: // bool
			v, err = MapOther(r, extra)
		default:
			panic(fmt.Errorf("unhandled major type: %d", valMaj))
		}
		if err != nil {
			return m, fmt.Errorf("error converting value of key %s: %w", key, err)
		}
		m[key] = v
	}

	return m, nil
}

func MapUInt(r *cbg.CborReader, extra uint64) (interface{}, error) {
	v := int64(extra)
	if v < 0 {
		return nil, fmt.Errorf("int64 positive overflow")
	}
	return v, nil
}

func MapInt(r *cbg.CborReader, extra uint64) (interface{}, error) {
	v := int64(extra)
	if v < 0 {
		return nil, fmt.Errorf("int64 negative oveflow")
	}
	v = -1 - v
	return v, nil
}

func MapByteString(r *cbg.CborReader, extra uint64) (interface{}, error) {
	b := make([]byte, extra)
	_, err := r.Read(b)
	if err != nil {
		return nil, fmt.Errorf("error reading byte string: %s", err)
	}
	return b, nil
}

func MapTextString(r *cbg.CborReader, extra uint64) (interface{}, error) {
	b := make([]byte, extra)
	_, err := r.Read(b)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func MapArray(r *cbg.CborReader, extra uint64) (interface{}, error) {
	vv := []interface{}{}
	var v interface{}
	for i := uint64(0); i < extra; i++ {
		valMaj, extra, err := r.ReadHeader()
		if err != nil {
			return nil, fmt.Errorf("error reading value header: %s", err)
		}
		switch valMaj {
		case cbg.MajMap:
			v, err = documentMap(r, extra)
		case cbg.MajUnsignedInt:
			v, err = MapUInt(r, extra)
		case cbg.MajNegativeInt:
			v, err = MapInt(r, extra)
		case cbg.MajByteString:
			v, err = MapByteString(r, extra)
		case cbg.MajTextString:
			v, err = MapTextString(r, extra)
		case cbg.MajArray:
			panic("nested arrays not supported")
		case cbg.MajTag:
			panic("arrays of tags not supported")
		case cbg.MajOther: // bool
			panic("arrays of bools not supported")
		default:
			panic("unhandled major type, " + fmt.Sprintf("%d", valMaj))
		}
		if err != nil {
			return nil, err
		}
		vv = append(vv, v)
	}
	return vv, nil
}

func MapTag(r *cbg.CborReader, extra uint64) (interface{}, error) {
	panic("MapTag not implemented")
}

func MapOther(r *cbg.CborReader, extra uint64) (interface{}, error) {
	switch extra {
	case 20: // false
		return false, nil
	case 21: // true
		return true, nil
	case 22: // null
		return nil, nil
	default:
		return nil, fmt.Errorf("booleans are either major type 7, value 20 or 21 (got %d)", extra)
	}
}
