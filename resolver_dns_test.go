//go:build integration
// +build integration

package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverDNS_E2E(t *testing.T) {
	exp0, err := ParsePeerAddr(
		`nimona://peer:addr:` +
			`2XxFa8qpbW4yV42XxFa8qpbW4yV4hEBwFYCyfsqQ21AupRBtXzWTzYaiNz` +
			`@utp:banks.testing.reamde.dev:1013`,
	)
	require.NoError(t, err)

	exp1, err := ParsePeerAddr(
		`nimona://peer:addr:` +
			`2XxFa8qpbW4yV4CYKa9qa42h5Nakx3Y5brfCqZZGZzMxvhzVG7YwyAfcY6` +
			`@utp:asimov.testing.reamde.dev:1013`,
	)
	require.NoError(t, err)

	exp := []PeerAddr{
		*exp0,
		*exp1,
	}

	nID, err := ParseNetworkAlias("nimona://net:alias:testing.reamde.dev")
	require.NoError(t, err)

	res := ResolverDNS{}
	addrs, err := res.Resolve(nID)
	require.NoError(t, err)
	require.Len(t, addrs, 2)
	require.Equal(t, exp, addrs)
}
