package nimona

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"
)

func TestPeerAddr(t *testing.T) {
	a := &PeerAddr{
		Network: "utp",
		Address: "localhost:1234",
	}
	t.Run("struct addr to string", func(t *testing.T) {
		require.Equal(t, "utp", a.Network)
		require.Equal(t, ResourceTypePeerAddress.String()+"utp:localhost:1234", a.String())
	})

	t.Run("string addr to struct", func(t *testing.T) {
		g, err := ParsePeerAddr(a.String())
		require.NoError(t, err)
		require.Equal(t, a, g)
	})

	t.Run("string addr to struct with public key", func(t *testing.T) {
		pub, _, err := ed25519.GenerateKey(nil)
		require.NoError(t, err)

		a = &PeerAddr{
			Network:   "utp",
			Address:   "localhost:1234",
			PublicKey: pub,
		}

		g, err := ParsePeerAddr(a.String())
		require.NoError(t, err)
		require.Equal(t, a, g)
	})
}

func TestPeerAddr_Marshal(t *testing.T) {
	pk, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	na := &PeerAddr{
		Network:   "utp",
		Address:   "localhost:1234",
		PublicKey: pk,
	}

	cb := new(bytes.Buffer)
	err = na.MarshalCBOR(cb)
	require.NoError(t, err)

	fmt.Printf("%x\n", cb.Bytes())

	ga := &PeerAddr{}
	err = ga.UnmarshalCBOR(cb)
	require.NoError(t, err)

	require.Equal(t, na, ga)
}
