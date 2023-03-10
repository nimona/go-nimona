package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverHTTP_ResolveIdentityAlias_E2E(t *testing.T) {
	expIdentityShort := "nimona://id:2ySDAozmRvfceoHQB1wcABQTzBYfaFydsprLEDPasGP4"

	expKeyGraphID, err := ParseDocumentID("nimona://doc:2ySDAozmRvfceoHQB1wcABQTzBYfaFydsprLEDPasGP4")
	require.NoError(t, err)

	expAsimovPublicKey, err := ParsePublicKey("9BfGM1GnGZXRvqVdpJUamdTLrmziAbdA8QZB1eMVYmoi")
	require.NoError(t, err)

	expBanksPublicKey, err := ParsePublicKey("CyN3L7rmCu5wnkCAcAMWv9iD5n2gJLaeLz6wpK6z31j9")
	require.NoError(t, err)

	alias, err := ParseIdentityAlias("nimona://id:alias:nimona.dev")
	require.NoError(t, err)

	exp := &IdentityInfo{
		Alias: IdentityAlias{
			Hostname: "nimona.dev",
		},
		Identity: Identity{
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

	res := ResolverHTTP{}
	got, err := res.ResolveIdentityAlias(*alias)
	require.NoError(t, err)
	require.Equal(t, exp, got)
	require.Equal(t, expIdentityShort, got.Identity.String())
}
