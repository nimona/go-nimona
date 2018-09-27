package primitives

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
