package net

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

func TestNetDiscoverer(t *testing.T) {
	_, n1 := newPeer(t, "")
	_, n2 := newPeer(t, "")

	n1.Discoverer().Add(n2.GetPeerInfo())
	n2.Discoverer().Add(n1.GetPeerInfo())

	q1 := &peer.PeerInfoRequest{SignerKeyHash: n2.key.GetPublicKey().HashBase58()}
	ps2, err := n1.Discoverer().Discover(q1)
	p2 := ps2[0]
	assert.NoError(t, err)
	// assert.Equal(t, n2.key.GetPublicKey(), p2.SignerKey)
	assert.Equal(t, n2.key.GetPublicKey().HashBase58(), p2.SignerKey.GetPublicKey().HashBase58())

	q2 := &peer.PeerInfoRequest{SignerKeyHash: n1.key.GetPublicKey().HashBase58()}
	ps1, err := n2.Discoverer().Discover(q2)
	p1 := ps1[0]
	assert.NoError(t, err)
	// assert.Equal(t, n1.key.GetPublicKey(), p1.SignerKey)
	assert.Equal(t, n1.key.GetPublicKey().HashBase58(), p1.SignerKey.HashBase58())
}

func newPeer(t *testing.T, relayAddress string) (*crypto.Key, *network) {
	tp, err := ioutil.TempDir("", "nimona-test-net")
	assert.NoError(t, err)

	kp := filepath.Join(tp, "key.cbor")

	pk, err := crypto.LoadKey(kp)
	assert.NoError(t, err)

	relayAddresses := []string{}
	if relayAddress != "" {
		relayAddresses = append(relayAddresses, relayAddress)
	}
	n, err := New(pk, "", relayAddresses)
	assert.NoError(t, err)

	return pk, n.(*network)
}
