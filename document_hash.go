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

	"nimona.io/internal/tilde"
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
func documentHashRaw(h tilde.Hint, b []byte) []byte {
	d := sha256.New()
	d.Write([]byte(string(h)))
	d.Write(b)
	return d.Sum(nil)
}

func NewDocumentHash(dm *DocumentMap) (h DocumentHash) {
	x, err := documentHashMap(dm.m)
	if err != nil {
		panic(fmt.Errorf("error hashing map: %w", err))
	}
	copy(h[:], x)
	return
}

func documentHashAny(valueAny tilde.Value) (h []byte, err error) {
	switch value := valueAny.(type) {
	case tilde.Ref:
		return []byte(value), nil
	case tilde.Uint64:
		return documentHashRaw(tilde.HintUint64, []byte(fmt.Sprintf("%d", value))), nil
	case tilde.Int64:
		return documentHashRaw(tilde.HintInt64, []byte(fmt.Sprintf("%d", value))), nil
	case tilde.Bytes:
		return documentHashRaw(tilde.HintBytes, value), nil
	case tilde.String:
		return documentHashRaw(tilde.HintString, []byte(value)), nil
	case tilde.Bool:
		if value {
			return documentHashRaw(tilde.HintBool, []byte{1}), nil
		}
		return documentHashRaw(tilde.HintBool, []byte{0}), nil
	case tilde.Map:
		if len(value) == 0 {
			return h, errDocumentHashValueIsNil
		}
		return documentHashMap(value)
	case tilde.List:
		if len(value) == 0 {
			return h, errDocumentHashValueIsNil
		}
		hr := new(bytes.Buffer)
		for _, value := range value {
			if value == nil {
				continue
			}
			hh, err := documentHashAny(value)
			if errors.Is(err, errDocumentHashValueIsNil) {
				continue
			}
			if err != nil {
				return h, fmt.Errorf("error hashing list value: %w", err)
			}
			hr.Write(hh)
		}
		return documentHashRaw(tilde.HintList, hr.Bytes()), nil
	default:
		panic(fmt.Errorf("unhandled type: %T", valueAny))
	}
}

type DocumentHashEntry struct {
	k     string
	khash []byte
	vhash []byte
}

type byKDocumentHash []DocumentHashEntry

func (h byKDocumentHash) Len() int      { return len(h) }
func (h byKDocumentHash) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h byKDocumentHash) Less(i, j int) bool {
	return bytes.Compare(h[i].khash, h[j].khash) < 0
}

func documentHashMap(m tilde.Map) (h []byte, err error) {
	e := byKDocumentHash{}
	for key, value := range m {
		// skip ephemeral fields
		if strings.HasPrefix(key, "_") {
			continue
		}
		// skip zero values
		// TODO(geoah): for all types?
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
			khash: documentHashRaw(tilde.HintString, []byte(key)),
			vhash: hh,
		})
	}

	sort.Sort(e)
	hr := new(bytes.Buffer)
	for _, ee := range e {
		hr.Write(ee.khash)
		hr.Write(ee.vhash)
	}
	return documentHashRaw(tilde.HintMap, hr.Bytes()), nil
}
