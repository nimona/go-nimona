package dht

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/codec"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

func TestPeerInfoResponseBlock(t *testing.T) {
	p1 := &peers.PeerInfo{
		Addresses: []string{
			"p1-addr1",
			"p1-addr2",
		},
	}

	p2 := &peers.PeerInfo{
		Addresses: []string{
			"p2-addr1",
			"p2-addr2",
		},
	}

	ep := &PeerInfoResponse{
		RequestID: "request-id",
		PeerInfo:  p1,
		ClosestPeers: []*peers.PeerInfo{
			p1,
			p2,
		},
	}

	b := ep.Block()
	bs, _ := codec.Marshal(b)

	b2 := &primitives.Block{}
	codec.Unmarshal(bs, b2)

	p := &PeerInfoResponse{}
	p.FromBlock(b2)

	assert.Equal(t, ep, p)
}
