package primitives // import "nimona.io/go/primitives"

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/base58"
)

func TestBlockExt(t *testing.T) {
	eb := &Block{
		Type: "block",
		Payload: map[string]interface{}{
			"alg": "key-alg",
		},
		Signature: &Signature{
			Key: &Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	bs, err := Marshal(eb)
	assert.NoError(t, err)

	fmt.Println(base58.Encode(bs))

	b, err := Unmarshal(bs)
	assert.NoError(t, err)
	assert.Equal(t, eb, b)
}
