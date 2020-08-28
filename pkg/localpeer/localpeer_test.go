package localpeer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func Test_memoryLocalPeer_GetPrimaryPeerKey(t *testing.T) {
	k1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	tests := []struct {
		key  crypto.PrivateKey
		want crypto.PrivateKey
	}{
		{
			key:  k1,
			want: k1,
		},
		{
			key:  k1,
			want: k1,
		},
		{
			key:  k2,
			want: k2,
		},
	}

	s := New()
	for _, tt := range tests {
		s.PutPrimaryPeerKey(tt.key)
		require.Equal(t, tt.want, s.GetPrimaryPeerKey())
	}
}

func Test_memoryLocalPeer_GetPrimaryIdentityKey(t *testing.T) {
	k1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	tests := []struct {
		key  crypto.PrivateKey
		want crypto.PrivateKey
	}{
		{
			key:  k1,
			want: k1,
		},
		{
			key:  k1,
			want: k1,
		},
		{
			key:  k2,
			want: k2,
		},
	}

	s := New()
	for _, tt := range tests {
		s.PutPrimaryIdentityKey(tt.key)
		require.Equal(t, tt.want, s.GetPrimaryIdentityKey())
	}
}
