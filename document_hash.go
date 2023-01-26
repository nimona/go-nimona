package nimona

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mr-tron/base58"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/exp/slices"
)

var errDocumentHashValueIsNil = fmt.Errorf("value is nil")

const hashLength = 32

type DocumentHash [hashLength]byte

func (h DocumentHash) String() string {
	return base58.Encode(h[:])
}

func (h DocumentHash) IsEqual(other DocumentHash) bool {
	return bytes.Equal(h[:], other[:])
}

func ParseDocumentHash(s string) (DocumentHash, error) {
	var h DocumentHash
	b, err := base58.Decode(s)
	if err != nil {
		return h, err
	}
	copy(h[:], b)
	return h, nil
}

// documentHashRaw hash the given value using sha256, prepending the given type
func documentHashRaw(t string, b []byte) [hashLength]byte {
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

func NewDocumentHash(c Cborer) (h DocumentHash, err error) {
	b, err := c.MarshalCBORBytes()
	if err != nil {
		return h, fmt.Errorf("error marshaling cbor: %s", err)
	}

	return NewDocumentHashFromCBOR(b)
}

func NewDocumentHashFromCBOR(b []byte) (h DocumentHash, err error) {
	r := cbg.NewCborReader(bytes.NewReader(b))
	maj, n, err := r.ReadHeader()
	if err != nil {
		return h, fmt.Errorf("error reading header: %s", err)
	}

	if maj != cbg.MajMap {
		return h, fmt.Errorf("cannot hash non maps")
	}

	h, err = documentHashMap(r, n)
	if err != nil {
		return h, fmt.Errorf("error hashing map: %s", err)
	}

	return h, nil
}

type DocumentHashEntry struct {
	khash [hashLength]byte
	vhash [hashLength]byte
}

type byKDocumentHash []DocumentHashEntry

func (h byKDocumentHash) Len() int      { return len(h) }
func (h byKDocumentHash) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h byKDocumentHash) Less(i, j int) bool {
	return bytes.Compare(h[i].khash[:],
		h[j].khash[:]) < 0
}

func documentHashMap(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	e := byKDocumentHash{}
	for i := uint64(0); i < extra; i++ {
		// read the key
		key, err := cbg.ReadString(r)
		if err != nil {
			return h, fmt.Errorf("error reading key: %w", err)
		}
		// skip ephemeral fields
		if strings.HasPrefix(key, "_") {
			continue
		}
		// read the value
		hh, err := documentHashAny(r)
		if errors.Is(err, errDocumentHashValueIsNil) {
			continue
		}
		if err != nil {
			return h, fmt.Errorf("error hashing value of key %s: %w", key, err)
		}
		e = append(e, DocumentHashEntry{
			khash: documentHashRaw("s", []byte(key)),
			vhash: hh,
		})
	}

	sort.Sort(e)
	hr := new(bytes.Buffer)
	for _, ee := range e {
		hr.Write(ee.khash[:])
		hr.Write(ee.vhash[:])
	}
	return documentHashRaw("m", hr.Bytes()), nil
}

func documentHashUInt(extra uint64) (h [hashLength]byte, err error) {
	extraI := int64(extra)
	if extraI < 0 {
		return h, fmt.Errorf("int64 positive overflow")
	}
	return documentHashRaw("u", []byte(fmt.Sprintf("%d", extraI))), nil
}

func documentHashInt(extra uint64) (h [hashLength]byte, err error) {
	extraI := int64(extra)
	if extraI < 0 {
		return h, fmt.Errorf("int64 negative oveflow")
	}
	extraI = -1 - extraI
	return documentHashRaw("i", []byte(fmt.Sprintf("%d", extraI))), nil
}

func documentHashByteString(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	b := make([]byte, extra)
	_, err = r.Read(b)
	if err != nil {
		return h, fmt.Errorf("error reading byte string: %s", err)
	}
	return documentHashRaw("d", b), nil
}

func documentHashTextString(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	b := make([]byte, extra)
	_, err = r.Read(b)
	if err != nil {
		return h, err
	}
	return documentHashRaw("s", b), nil
}

func documentHashAny(r *cbg.CborReader) (h [hashLength]byte, err error) {
	maj, extra, err := r.ReadHeader()
	if err != nil {
		return h, err
	}
	switch maj {
	case cbg.MajMap:
		return documentHashMap(r, extra)
	case cbg.MajUnsignedInt:
		return documentHashUInt(extra)
	case cbg.MajNegativeInt:
		return documentHashInt(extra)
	case cbg.MajByteString:
		if extra == 0 {
			return h, errDocumentHashValueIsNil
		}
		return documentHashByteString(r, extra)
	case cbg.MajTextString:
		if extra == 0 {
			return h, errDocumentHashValueIsNil
		}
		return documentHashTextString(r, extra)
	case cbg.MajArray:
		if extra == 0 {
			return h, errDocumentHashValueIsNil
		}
		return documentHashArray(r, extra)
	case cbg.MajTag:
		panic("tags not supported")
	case cbg.MajOther: // bool
		return documentHashOther(extra)
	default:
		panic(fmt.Errorf("unhandled major type: %d", maj))
	}
}

func documentHashArray(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	hr := new(bytes.Buffer)
	for i := uint64(0); i < extra; i++ {
		hh, err := documentHashAny(r)
		if errors.Is(err, errDocumentHashValueIsNil) {
			continue
		}
		if err != nil {
			return h, err
		}
		hr.Write(hh[:])
	}
	return documentHashRaw("a", hr.Bytes()), nil
}

// nolint:unused,deadcode // TODO: implement
func documentHashTag(r *cbg.CborReader, extra uint64) (h [hashLength]byte, err error) {
	panic("documentHashTag not implemented")
}

func documentHashOther(extra uint64) (h [hashLength]byte, err error) {
	switch extra {
	case 20: // false
		return documentHashRaw("b", []byte{0}), nil
	case 21: // true
		return documentHashRaw("b", []byte{1}), nil
	case 22: // null
		return documentHashRaw("b", []byte{0}), nil
	default:
		return h, fmt.Errorf("booleans are either major type 7, value 20 or 21 (got %d)", extra)
	}
}
