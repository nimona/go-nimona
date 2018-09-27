package dht

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/codec"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

func TestProviderResponseBlock(t *testing.T) {
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

	pr1 := &Provider{
		BlockIDs: []string{
			"p1-prov1",
			"p1-prov2",
		},
	}

	pr2 := &Provider{
		BlockIDs: []string{
			"p2-prov1",
			"p2-prov2",
		},
	}

	ep := &ProviderResponse{
		RequestID: "request-id",
		Providers: []*Provider{
			pr1,
			pr2,
		},
		ClosestPeers: []*peers.PeerInfo{
			p1,
			p2,
		},
	}

	b := ep.Block()
	bs, _ := codec.Marshal(b)

	b2 := &primitives.Block{}
	codec.Unmarshal(bs, b2)

	p := &ProviderResponse{}
	p.FromBlock(b2)

	assert.Equal(t, ep, p)
}
