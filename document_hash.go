package nimona

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mr-tron/base58"
	"github.com/vikyd/zero"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/exp/slices"
)

var errDocumentHashValueIsNil = fmt.Errorf("value is nil")

const hashLength = 32

type (
	DocumentHash [hashLength]byte
)

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

func NewDocumentHash(dm DocumentMapper) (h DocumentHash) {
	var err error
	if m, ok := dm.(DocumentMap); ok {
		h, err = documentHashMap(m)
	} else {
		h, err = documentHashMap(dm.DocumentMap())
	}
	if err != nil {
		panic(fmt.Errorf("error hashing map: %w", err))
	}
	return h
}

type DocumentHashEntry struct {
	k     string
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

func documentHashMap(m DocumentMap) (h [hashLength]byte, err error) {
	e := byKDocumentHash{}
	for key, value := range m {
		// skip ephemeral fields
		if strings.HasPrefix(key, "_") {
			continue
		}
		// skip zero values // TODO(geoah): for all types?
		if zero.IsZeroVal(value) {
			continue
		}
		// hash the value
		hh, err := documentHashAny(value)
		if errors.Is(err, errDocumentHashValueIsNil) {
			continue
		}
		if err != nil {
			return h, fmt.Errorf("error hashing value of key %s: %w", key, err)
		}
		e = append(e, DocumentHashEntry{
			k:     key,
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

func documentHashUInt(value uint64) (h [hashLength]byte) {
	return documentHashRaw("u", []byte(fmt.Sprintf("%d", value)))
}

func documentHashInt(value int64) (h [hashLength]byte) {
	return documentHashRaw("i", []byte(fmt.Sprintf("%d", value)))
}

func documentHashBytes(value []byte) (h [hashLength]byte) {
	return documentHashRaw("d", value)
}

func documentHashTextString(value string) (h [hashLength]byte) {
	return documentHashRaw("s", []byte(value))
}

func documentHashAny(valueAny any) (h [hashLength]byte, err error) {
	switch value := valueAny.(type) {
	case DocumentHash:
		return value, nil
	case DocumentMap:
		return documentHashMap(value)
	case map[string]interface{}:
		return documentHashMap(DocumentMap(value))
	case uint64:
		return documentHashUInt(value), nil
	case int64:
		return documentHashInt(value), nil
	case []byte:
		return documentHashBytes(value), nil
	case string:
		return documentHashTextString(value), nil
	case bool:
		return documentHashBool(value), nil
	case []int64:
		return documentHashArray(value)
	case []uint64:
		return documentHashArray(value)
	case []string:
		return documentHashArray(value)
	case []bool:
		return documentHashArray(value)
	case [][]byte:
		return documentHashArray(value)
	case []any:
		return documentHashArray(value)
	default:
		panic(fmt.Errorf("unhandled type: %T", valueAny))
	}
}

func documentHashArray[E any](values []E) (h [hashLength]byte, err error) {
	if len(values) == 0 {
		return h, errDocumentHashValueIsNil
	}
	hr := new(bytes.Buffer)
	for _, value := range values {
		hh, err := documentHashAny(value)
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

func documentHashBool(value bool) (h [hashLength]byte) {
	if value {
		return documentHashRaw("b", []byte{1})
	}
	return documentHashRaw("b", []byte{0})
}
