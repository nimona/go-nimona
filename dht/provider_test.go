package dht

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/base58"
	"nimona.io/go/codec"
	"nimona.io/go/primitives"
)

func TestProviderBlock(t *testing.T) {
	ep := &Provider{
		BlockIDs: []string{
			"block1",
			"block2",
		},
		Signature: &primitives.Signature{
			Key: &primitives.Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	b := ep.Block()
	bs, _ := primitives.Marshal(b)

	b2, err := primitives.Unmarshal(bs)
	assert.NoError(t, err)

	p := &Provider{}
	p.FromBlock(b2)

	assert.Equal(t, ep, p)
}

func TestProviderSelfEncode(t *testing.T) {
	eb := &Provider{
		BlockIDs: []string{
			"block1",
			"block2",
		},
		Signature: &primitives.Signature{
			Key: &primitives.Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	bs, err := codec.Marshal(eb)
	assert.NoError(t, err)

	assert.Equal(t, "ZWDuV4VZZGrf2q1AeWbQLDJQP6BbKsYbUPBxAEwR9cc3sb2YG23sNawHv"+
		"2HXMkmDxVVWkRC2g9oNempjhioM3x2iaogHB6rT4msUza7Es82frJk64rzfyTdEJNojA8"+
		"mPrLpdCCAbAZCRvtj2Myc18crE1AKa2c1xmndymR8a4n9eEfEqY7DHGCpUt927mRSbpRB"+
		"Xq1qXihTJbpPCGq8mBm1A6W9NcDAq", base58.Encode(bs))

	b := &Provider{}
	err = codec.Unmarshal(bs, b)
	assert.NoError(t, err)

	assert.Equal(t, eb, b)
}
