package object

import (
	"strings"

	"nimona.io/internal/encoding/base58"
)

type (
	Hash string
)

func (h Hash) IsEmpty() bool {
	return h == ""
}

func (h Hash) IsEqual(c Hash) bool {
	return h == c
}

func (h Hash) String() string {
	return string(h)
}

func (h Hash) Bytes() []byte {
	ps := strings.Split(string(h), ".")
	if len(ps) == 0 {
		return nil
	}
	p := ps[len(ps)-1]
	b, _ := base58.Decode(p)
	return b
}
