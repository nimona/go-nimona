package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeerConfig(t *testing.T) {
	pc := NewTestPeerConfig(t)

	// Test GetPrivateKey()
	require.True(t, pc.GetPrivateKey().Equal(pc.privateKey))

	// Test GetPublicKey()
	require.True(t, pc.GetPublicKey().Equal(pc.publicKey))

	// Test GetPeerInfo()
	require.EqualValues(t, pc.peerInfo, pc.GetPeerInfo())

	// Test GetIdentity().IdentityID() when identity is nil
	pc.identity = nil
	require.Nil(t, pc.GetIdentity())
}

func NewTestPeerConfig(t *testing.T) *PeerConfig {
	t.Helper()

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

	return pc
}
