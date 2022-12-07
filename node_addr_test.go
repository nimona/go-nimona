package nimona

import (
	"testing"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"
)

func TestNodeAddr(t *testing.T) {
	a := NodeAddr{
		Host:      "localhost",
		Port:      1234,
		Transport: "utp",
	}

	t.Run("struct addr to string", func(t *testing.T) {
		require.Equal(t, "utp", a.Network())
		require.Equal(t, "localhost:1234", a.Address())
		require.Equal(t, PeerAddressPrefix+"utp:localhost:1234", a.String())
	})

	t.Run("string addr to struct", func(t *testing.T) {
		g := NodeAddr{}
		err := g.Parse(a.String())
		require.NoError(t, err)
		require.Equal(t, a, g)
	})

	t.Run("string addr to struct with public key", func(t *testing.T) {
		pub, _, err := ed25519.GenerateKey(nil)
		require.NoError(t, err)
		a.PublicKey = pub

		g := NodeAddr{}
		err = g.Parse(a.String())
		require.NoError(t, err)
		require.Equal(t, a, g)
	})
}
