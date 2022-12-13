package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverDNS_E2E(t *testing.T) {
	exp := []NodeAddr{{}, {}}
	exp[0].Parse(
		`nimona://peer:addr:` +
			`2XxFa8qpbW4yV42XxFa8qpbW4yV4hEBwFYCyfsqQ21AupRBtXzWTzYaiNz` +
			`@utp:banks.testing.reamde.dev:1013`,
	)
	exp[1].Parse(
		`nimona://peer:addr:` +
			`2XxFa8qpbW4yV4CYKa9qa42h5Nakx3Y5brfCqZZGZzMxvhzVG7YwyAfcY6` +
			`@utp:asimov.testing.reamde.dev:1013`,
	)

	res := ResolverDNS{}
	addrs, err := res.Resolve("nimona://peer:handle:testing.reamde.dev")
	require.NoError(t, err)
	require.Len(t, addrs, 2)
	require.Equal(t, exp, addrs)
}
