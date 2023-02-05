package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeerAddr(t *testing.T) {
	a := &PeerAddr{
		Network: "utp",
		Address: "localhost:1234",
	}
	t.Run("struct addr to string", func(t *testing.T) {
		require.Equal(t, "utp", a.Network)
		require.Equal(t, ShorthandPeerAddress.String()+"utp:localhost:1234", a.String())
	})

	t.Run("string addr to struct", func(t *testing.T) {
		g, err := ParsePeerAddr(a.String())
		require.NoError(t, err)
		require.Equal(t, a, g)
	})

	t.Run("string addr to struct with public key", func(t *testing.T) {
		pub, _, err := GenerateKey()
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
