package nimona

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverHTTP_ResolveIdentityAlias_E2E(t *testing.T) {
	expKeyGraphID, err := ParseKeyGraphNRI("nimona://id:EbEZ8YXH4YvLrgG4mHCutfEi8oYhZjPrdVQZLiUa74yJ")
	require.NoError(t, err)

	expAsimovPublicKey, err := ParsePublicKey("94vLdxU9gpvmsJo1N3GXjbXqRDS53Bkeh4yPyt6cmH71")
	require.NoError(t, err)

	expBanksPublicKey, err := ParsePublicKey("4GNhCLJ7AMjMHEGNY5YwbVNReTyjtMpoF5E55rB2iPr3")
	require.NoError(t, err)

	alias, err := ParseIdentityAlias("nimona://id:alias:nimona.dev")
	require.NoError(t, err)

	exp := &IdentityInfo{
		Alias: IdentityAlias{
			Hostname: "nimona.dev",
		},
		KeyGraphID: expKeyGraphID,
		PeerAddresses: []PeerAddr{{
			Address:   "asimov.testing.reamde.dev:1013",
			Transport: "utp",
			PublicKey: expAsimovPublicKey,
		}, {
			Address:   "banks.testing.reamde.dev:1013",
			Transport: "utp",
			PublicKey: expBanksPublicKey,
		}},
	}

	res := ResolverHTTP{}
	got, err := res.ResolveIdentityAlias(*alias)

	DumpDocument(got.Document())

	fmt.Println(got.KeyGraphID.String())
	fmt.Println(got.PeerAddresses[0].PublicKey.String())
	fmt.Println(got.PeerAddresses[1].PublicKey.String())

	require.NoError(t, err)
	require.Equal(t, exp, got)
	require.Equal(t, expKeyGraphID, got.KeyGraphID)
}
