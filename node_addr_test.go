package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeAddr(t *testing.T) {
	a := NewNodeAddr("utp", "localhost", 1234)

	t.Run("struct addr to string", func(t *testing.T) {
		require.Equal(t, "utp", a.Network())
		require.Equal(t, "localhost:1234", a.Address())
		require.Equal(t, "nimona://utp:localhost:1234", a.String())
	})

	t.Run("string addr to struct", func(t *testing.T) {
		g := NodeAddr{}
		err := g.Parse(a.String())
		require.NoError(t, err)
		require.Equal(t, a, g)
	})

	t.Run("string addr to struct with extras", func(t *testing.T) {
		g := NodeAddr{}
		err := g.Parse("nimona://utp:localhost:1234/something")
		require.NoError(t, err)
		require.Equal(t, NodeAddr{
			host:      "localhost",
			port:      1234,
			transport: "utp",
			extra:     "something",
		}, g)
	})
}
