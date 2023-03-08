package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverHTTP_ResolveIdentityAlias_E2E(t *testing.T) {
	expKeyGraphID, err := ParseDocumentID("nimona://doc:316dLmTi6NLp13rCPv6E5WDZKjFqKrp8sNvxefLYrRro")
	require.NoError(t, err)

	expAsimovPublicKey, err := ParsePublicKey("Au4XYAngGMB1Zj8mVJifzH2FgUJb8TVYLuN8UDUW6giz")
	require.NoError(t, err)

	expBanksPublicKey, err := ParsePublicKey("F9CiQ9XzWGr3cfAp2wGmavkXU5TwWPDxH6Ck39J4bdRb")
	require.NoError(t, err)

	exp := &NetworkInfo{
		NetworkAlias: NetworkAlias{
			Hostname: "nimona.io",
		},
		NetworkIdentity: Identity{
			KeyGraphID: expKeyGraphID,
		},
		PeerAddresses: []PeerAddr{{
			Address:   "asimov.testing.reamde.dev:1013",
			Network:   "utp",
			PublicKey: expAsimovPublicKey,
		}, {
			Address:   "banks.testing.reamde.dev:1013",
			Network:   "utp",
			PublicKey: expBanksPublicKey,
		}},
	}
	alias, err := ParseIdentityAlias("nimona://id:alias:nimona.dev")
	require.NoError(t, err)

	res := ResolverHTTP{}
	got, err := res.ResolveIdentityAlias(alias)
	require.NoError(t, err)
	require.Equal(t, exp, got)
}
