package dht // import "nimona.io/go/dht"

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

func TestPeerInfoResponseBlock(t *testing.T) {
	p1 := &peers.PeerInfo{
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

	p2 := &peers.PeerInfo{
		Addresses: []string{
			"p2-addr1",
			"p2-addr2",
		},
		Signature: &primitives.Signature{
			Key: &primitives.Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	ep := &PeerInfoResponse{
		RequestID: "request-id",
		PeerInfo:  p1,
		ClosestPeers: []*peers.PeerInfo{
			p1,
			p2,
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

	p := &PeerInfoResponse{}
	p.FromBlock(b2)

	assert.Equal(t, ep, p)
}
