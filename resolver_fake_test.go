package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverFake_ResolveIdentityAlias_E2E(t *testing.T) {
	expKeyGraphID, err := ParseDocumentID("nimona://doc:2ySDAozmRvfceoHQB1wcABQTzBYfaFydsprLEDPasGP4")
	require.NoError(t, err)

	expAsimovPublicKey, err := ParsePublicKey("9BfGM1GnGZXRvqVdpJUamdTLrmziAbdA8QZB1eMVYmoi")
	require.NoError(t, err)

	expBanksPublicKey, err := ParsePublicKey("CyN3L7rmCu5wnkCAcAMWv9iD5n2gJLaeLz6wpK6z31j9")
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
	alias, err := ParseIdentityAlias("nimona://id:alias:nimona.dev")
	require.NoError(t, err)

	res := ResolverFake{
		identities: map[string]*IdentityInfo{
			"nimona.dev": exp,
		},
	}
	got, err := res.ResolveIdentityAlias(*alias)
	require.NoError(t, err)
	require.Equal(t, exp, got)
}
