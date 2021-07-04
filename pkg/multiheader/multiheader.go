// Package multiheader generalizes the various multiformats by allowing adding
// and parsing a multicodec code to/from bytes without having or needing to care
// if the value is a cid/hash/addr/etc.

package multiheader

import (
	"fmt"

	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-varint"
)

type (
	Multiheader []byte
)

func Encode(code multicodec.Code, raw []byte) Multiheader {
	s := varint.UvarintSize(uint64(code)) + len(raw)
	b := make([]byte, s)
	n := varint.PutUvarint(b, uint64(code))
	copy(b[n:], raw)
	return b
}

func Decode(b []byte) (code multicodec.Code, raw []byte, err error) {
	c, n, err := varint.FromUvarint(b)
	if err != nil {
		return 0, nil, fmt.Errorf("unable to read uvarint, %w", err)
	}

	return multicodec.Code(c), b[n:], nil
}
