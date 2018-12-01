package crypto

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/base58"
	"nimona.io/go/encoding"
)

func TestSignatureEncoding(t *testing.T) {
	es := &Signature{
		Alg: "sig-alg",
		R:   big.NewInt(math.MaxInt64).Bytes(),
		S:   big.NewInt(math.MinInt64).Bytes(),
	}

	eo, err := encoding.NewObjectFromStruct(es)
	assert.NoError(t, err)

	bs, err := encoding.Marshal(eo)
	assert.NoError(t, err)

	assert.Equal(t, "41BGbraog8gf47JJunL1pYsE8eeE11v6uirNPoFNuKJXP92EZudmWzi19"+
		"mdGmoTnyMxwCWvSj", base58.Encode(bs))

	o, err := encoding.Unmarshal(bs)
	assert.NoError(t, err)
	assert.Equal(t, eo, o)

	s, err := o.Materialize()
	assert.Equal(t, es, s)
}
