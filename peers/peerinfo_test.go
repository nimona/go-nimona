package peers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/base58"
	"nimona.io/go/codec"
	"nimona.io/go/primitives"
)

func TestPeerInfoBlock(t *testing.T) {
	ep := &PeerInfo{
		Addresses: []string{
			"p1-addr1",
			"p1-addr2",
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

	p := &PeerInfo{}
	p.FromBlock(b2)

	assert.Equal(t, ep, p)
}

func TestPeerInfoSelfEncode(t *testing.T) {
	eb := &PeerInfo{
		Addresses: []string{
			"p1-addr1",
			"p1-addr2",
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

	assert.Equal(t, base58.Encode(bs), "BvE6Qe57DKXhLXzNVg4HeDf6Gv3jFAmZzixdtB"+
		"jLmkQQP9tBLgsRtCPRDF5gUnt4FXZuxMNNbJwScDHwgnRr1SZyK9fNv7zUpV2LyQPCGXk"+
		"wDQ5rGruw8bTfjvfyg9gQiPTWEH5JtCNocVJiEAqj9qrFYtmrKsVDibAL5EJ53dxZrb5M"+
		"UPArU2ze2Yy7jhpib1YxGNZv89WAACh9E4fRRbDQmWaoMf2BLiS")

	b := &PeerInfo{}
	err = codec.Unmarshal(bs, b)
	assert.NoError(t, err)

	assert.Equal(t, eb, b)
}
