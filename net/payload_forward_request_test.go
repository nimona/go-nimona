package net

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/base58"
	"nimona.io/go/primitives"
)

func TestForwardRequestBlock(t *testing.T) {
	ev := &ForwardRequest{
		Recipient: "foo",
		FwBlock: &primitives.Block{
			Type: "wrapped",
			Payload: map[string]interface{}{
				"foo": "bar",
			},
			Signature: &primitives.Signature{
				Key: &primitives.Key{
					Algorithm: "key-alg",
				},
				Alg: "sig-alg",
			},
		},
		Signature: &primitives.Signature{
			Key: &primitives.Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	eb := ev.Block()

	bs, err := primitives.Marshal(eb)
	assert.NoError(t, err)

	fmt.Println(base58.Encode(bs))

	b, err := primitives.Unmarshal(bs)
	assert.NoError(t, err)

	v := &ForwardRequest{}
	v.FromBlock(b)

	assert.Equal(t, ev, v)
}
