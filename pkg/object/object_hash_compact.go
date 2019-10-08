package object

import (
	"strings"

	"nimona.io/pkg/errors"

	"nimona.io/internal/encoding/base58"
)

func HashFromCompact(s string) (*Hash, error) {
	p := strings.Split(s, ".")
	if len(p) != 2 {
		return nil, errors.New("invalid compact hash")
	}
	b, err := base58.Decode(p[1])
	if err != nil {
		return nil, err
	}
	return &Hash{
		Algorithm: p[0],
		D:         b,
	}, nil
}

func (h Hash) Compact() string {
	return h.Algorithm + "." + base58.Encode(h.D)
}

func (h Hash) String() string {
	return h.Compact()
}
