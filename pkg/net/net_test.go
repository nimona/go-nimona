package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/pkg/discovery"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

func TestNetDiscoverer(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	_, n1 := newPeer(t, "", disc1)
	_, n2 := newPeer(t, "", disc2)

	disc1.Add(n2.GetPeerInfo())
	disc2.Add(n1.GetPeerInfo())

	q1 := &peer.PeerInfoRequest{SignerKeyHash: n2.key.GetPublicKey().HashBase58()}
	ps2, err := disc1.Discover(q1)
	p2 := ps2[0]
	assert.NoError(t, err)
	// assert.Equal(t, n2.key.GetPublicKey(), p2.SignerKey)
	assert.Equal(t, n2.key.GetPublicKey().HashBase58(), p2.SignerKey.GetPublicKey().HashBase58())

	q2 := &peer.PeerInfoRequest{SignerKeyHash: n1.key.GetPublicKey().HashBase58()}
	ps1, err := disc2.Discover(q2)
	p1 := ps1[0]
	assert.NoError(t, err)
	// assert.Equal(t, n1.key.GetPublicKey(), p1.SignerKey)
	assert.Equal(t, n1.key.GetPublicKey().HashBase58(), p1.SignerKey.HashBase58())
}

func newPeer(t *testing.T, relayAddress string,
	discover discovery.Discoverer) (*crypto.Key, *network) {
	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	relayAddresses := []string{}
	if relayAddress != "" {
		relayAddresses = append(relayAddresses, relayAddress)
	}
	n, err := New(pk, "", relayAddresses, discover)
	assert.NoError(t, err)

	return pk, n.(*network)
}
