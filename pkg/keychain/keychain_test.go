package keychain

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func Test_memoryKeychain_List(t *testing.T) {
	k1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	k3, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	tests := []struct {
		keytype KeyType
		key     crypto.PrivateKey
		want    []crypto.PrivateKey
	}{
		{
			keytype: PeerKey,
			key:     k1,
			want:    []crypto.PrivateKey{k1},
		},
		{
			keytype: PeerKey,
			key:     k2,
			want:    []crypto.PrivateKey{k1, k2},
		},
		{
			keytype: IdentityKey,
			key:     k3,
			want:    []crypto.PrivateKey{k3},
		},
		{
			keytype: PeerKey,
			key:     k2,
			want:    []crypto.PrivateKey{k1, k2},
		},
		{
			keytype: IdentityKey,
			key:     k1,
			want:    []crypto.PrivateKey{k3, k1},
		},
	}

	s := New()
	for _, tt := range tests {
		s.Put(tt.keytype, tt.key)
		require.ElementsMatch(t, tt.want, s.List(tt.keytype))
	}
}

func Test_memoryKeychain_GetPrimaryPeerKey(t *testing.T) {
	k1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	tests := []struct {
		keytype KeyType
		key     crypto.PrivateKey
		want    crypto.PrivateKey
	}{
		{
			keytype: PrimaryPeerKey,
			key:     k1,
			want:    k1,
		},
		{
			keytype: PrimaryPeerKey,
			key:     k1,
			want:    k1,
		},
		{
			keytype: PrimaryPeerKey,
			key:     k2,
			want:    k2,
		},
	}

	s := New()
	for _, tt := range tests {
		s.Put(tt.keytype, tt.key)
		require.Equal(t, tt.want, s.GetPrimaryPeerKey())
	}
}
