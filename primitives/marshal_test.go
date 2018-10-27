package primitives // import "nimona.io/go/primitives"

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	b := &Block{
		Type: "type",
		Payload: map[string]interface{}{
			"foo": "bar",
		},
		Signature: &Signature{
			Key: &Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	bs, err := Marshal(b)
	assert.NoError(t, err)
	assert.NotNil(t, bs)

	nb, err := Unmarshal(bs)
	assert.NoError(t, err)
	assert.NotNil(t, nb)

	assert.Equal(t, b, nb)
}
