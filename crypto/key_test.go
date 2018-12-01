package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/base58"
	"nimona.io/go/encoding"
)

func TestKeyEncoding(t *testing.T) {
	ek := &Key{
		Algorithm: "key-alg",
	}

	eo, err := encoding.NewObjectFromStruct(ek)
	assert.NoError(t, err)

	bs, err := encoding.Marshal(eo)
	assert.NoError(t, err)

	assert.Equal(t, "Nx3cnuT6J8XPNCBmncEt5BfwfKtYtf6h5VzoQ", base58.Encode(bs))

	o, err := encoding.Unmarshal(bs)
	assert.NoError(t, err)
	assert.Equal(t, eo, o)

	k, err := o.Materialize()
	assert.NoError(t, err)
	assert.Equal(t, ek, k)
}
