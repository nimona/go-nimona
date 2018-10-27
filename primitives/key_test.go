package primitives // import "nimona.io/go/primitives"

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/base58"
	"nimona.io/go/codec"
)

func TestKeyBlock(t *testing.T) {
	k := &Key{
		Algorithm: "key-alg",
	}

	eb := &Block{
		Type: "nimona.io/key",
		Payload: map[string]interface{}{
			"alg": "key-alg",
		},
	}

	b := k.Block()
	assert.Equal(t, eb, b)
}

func TestKeyFromBlock(t *testing.T) {
	ek := &Key{
		Algorithm: "key-alg",
	}

	b := &Block{
		Type: "nimona.io/key",
		Payload: map[string]interface{}{
			"alg": "key-alg",
		},
	}

	k := &Key{}
	k.FromBlock(b)

	assert.Equal(t, ek, k)
}

func TestKeySelfEncode(t *testing.T) {
	k := &Key{
		Algorithm: "key-alg",
	}

	bs, err := codec.Marshal(k)
	assert.NoError(t, err)

	fmt.Println(base58.Encode(bs))

	eb := &Block{
		Type: "nimona.io/key",
		Payload: map[string]interface{}{
			"alg": "key-alg",
		},
	}

	b := &Block{}
	err = codec.Unmarshal(bs, b)
	assert.NoError(t, err)

	assert.Equal(t, eb, b)
}
