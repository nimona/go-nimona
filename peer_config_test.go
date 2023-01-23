package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeerConfig(t *testing.T) {
	pk, sk, err := GenerateKey()
	require.NoError(t, err)

	// Create test private and public keys
	peerInfo := &PeerInfo{
		PublicKey: pk,
		Addresses: []PeerAddr{{
			Address:   "test:localhost",
			PublicKey: pk,
		}},
	}

	// Create a new PeerConfig
	pc := NewPeerConfig(sk, pk, peerInfo)

	// Test GetPrivateKey()
	require.True(t, pc.GetPrivateKey().Equal(sk))

	// Test GetPublicKey()
	require.True(t, pc.GetPublicKey().Equal(pk))

	// Test GetPeerInfo()
	require.EqualValues(t, peerInfo, pc.GetPeerInfo())
}
