package nimona

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/mr-tron/base58"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/exp/slices"
)

const hashLength = 32

type Hash [hashLength]byte

func (h Hash) String() string {
	return base58.Encode(h[:])
}

func HashFromBase58(s string) (Hash, error) {
	var h Hash
	b, err := base58.Decode(s)
	if err != nil {
		return h, err
	}
	copy(h[:], b)
	return h, nil
}

// messageHashRaw hash the given value using sha256, prepending the given type
func messageHashRaw(t string, b []byte) [hashLength]byte {
	if !slices.Contains([]string{
		"s", "i", "u", "b", "d", "a", "m",
	}, t) {
		panic(fmt.Errorf("invalid type: %s", t))
	}
	h := sha256.New()
	h.Write([]byte(t))
	h.Write(b)

	var r [hashLength]byte
	copy(r[:], h.Sum(nil))
	return r
}

func MessageHash(c Cborer) (h Hash, err error) {
	b, err := c.MarshalCBORBytes()
	if err != nil {
		return h, err
	}

	return MessageHashFromCBOR(b)
}

func MessageHashFromCBOR(b []byte) (h Hash, err error) {
	r := cbg.NewCborReader(bytes.NewReader(b))
	maj, n, err := r.ReadHeader()
	if err != nil {
		return h, err
	}

	if maj != cbg.MajMap {
		return h, fmt.Errorf("cannot hash non maps")
	}

	return messageHashMap(r, n)
}

type MessageHashEntry struct {
	khash [hashLength]byte
	vhash [hashLength]byte
}

type byKHash []MessageHashEntry

func (h byKHash) Len() int      { return len(h) }
func (h byKHash) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h byKHash) Less(i, j int) bool {
	return bytes.Compare(h[i].khash[:],
		h[j].khash[:]) < 0
}

func messageHashMap(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	e := byKHash{}
	for i := uint64(0); i < extra; i++ {
		key, err := cbg.ReadString(r)
		if err != nil {
			return h, err
		}
		valMaj, extra, err := r.ReadHeader()
		if err != nil {
			return h, fmt.Errorf("error reading value header: %s", err)
		}
		var hh [hashLength]byte
		switch valMaj {
		case cbg.MajMap:
			hh, err = messageHashMap(r, extra)
		case cbg.MajUnsignedInt:
			hh, err = messageHashUInt(r, extra)
		case cbg.MajNegativeInt:
			hh, err = messageHashInt(r, extra)
		case cbg.MajByteString:
			hh, err = messageHashByteString(r, extra)
		case cbg.MajTextString:
			hh, err = messageHashTextString(r, extra)
		case cbg.MajArray:
			hh, err = messageHashArray(r, extra)
		// case cbg.MajTag:
		// 	hh, err = messageHashTag(r, extra)
		case cbg.MajOther: // bool
			hh, err = messageHashOther(r, extra)
		default:
			panic(fmt.Errorf("unhandled major type: %d", valMaj))
		}
		if err != nil {
			return h, err
		}
		e = append(e, MessageHashEntry{
			khash: messageHashRaw("s", []byte(key)),
			vhash: hh,
		})
	}

	sort.Sort(byKHash(e))
	hr := new(bytes.Buffer)
	for _, ee := range e {
		hr.Write(ee.khash[:])
		hr.Write(ee.vhash[:])
	}
	return messageHashRaw("m", hr.Bytes()), nil
}

func messageHashUInt(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	extraI := int64(extra)
	if extraI < 0 {
		return h, fmt.Errorf("int64 positive overflow")
	}
	return messageHashRaw("u", []byte(fmt.Sprintf("%d", extraI))), nil
}

func messageHashInt(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	extraI := int64(extra)
	if extraI < 0 {
		return h, fmt.Errorf("int64 negative oveflow")
	}
	extraI = -1 - extraI
	return messageHashRaw("i", []byte(fmt.Sprintf("%d", extraI))), nil
}

func messageHashByteString(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	b := make([]byte, extra)
	_, err = r.Read(b)
	if err != nil {
		return h, err
	}
	return messageHashRaw("d", b), nil
}

func messageHashTextString(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	b := make([]byte, extra)
	_, err = r.Read(b)
	if err != nil {
		return h, err
	}
	return messageHashRaw("s", b), nil
}

func messageHashArray(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	hr := new(bytes.Buffer)
	for i := uint64(0); i < extra; i++ {
		valMaj, extra, err := r.ReadHeader()
		if err != nil {
			return h, fmt.Errorf("error reading value header: %s", err)
		}
		var hh [hashLength]byte
		switch valMaj {
		case cbg.MajMap:
			hh, err = messageHashMap(r, extra)
		case cbg.MajUnsignedInt:
			hh, err = messageHashUInt(r, extra)
		case cbg.MajNegativeInt:
			hh, err = messageHashInt(r, extra)
		case cbg.MajByteString:
			hh, err = messageHashByteString(r, extra)
		case cbg.MajTextString:
			hh, err = messageHashTextString(r, extra)
		// case cbg.MajArray:
		// 	hh, err = messageHashArray(r, extra)
		// case cbg.MajTag:
		// 	hh, err = messageHashTag(r, extra)
		// case cbg.MajOther: // bool
		// 	hh, err = messageHashOther(r, extra)
		default:
			panic("unhandled major type, " + fmt.Sprintf("%d", valMaj))
		}
		if err != nil {
			return h, err
		}
		hr.Write(hh[:])
	}
	return messageHashRaw("a", hr.Bytes()), nil
}

func messageHashTag(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	panic("messageHashTag not implemented")
}

func messageHashOther(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	switch extra {
	case 20: // false
		return messageHashRaw("b", []byte{0}), nil
	case 21: // true
		return messageHashRaw("b", []byte{1}), nil
	case 22: // null
		return messageHashRaw("b", []byte{0}), nil
	default:
		return h, fmt.Errorf("booleans are either major type 7, value 20 or 21 (got %d)", extra)
	}
}
