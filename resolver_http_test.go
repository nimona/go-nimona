package nimona

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverHTTP_ResolveIdentityAlias_E2E(t *testing.T) {
	expIdentityShort := "nimona://id:CFCPKiBrMMtdG7hwpFHv32jiAeLFKzu9ML3iuvMmGVkf"

	expKeyGraphID, err := ParseDocumentNRI("nimona://doc:CFCPKiBrMMtdG7hwpFHv32jiAeLFKzu9ML3iuvMmGVkf")
	require.NoError(t, err)

	expAsimovPublicKey, err := ParsePublicKey("AqYceSpfuEWsg6LNKqc1rPn232MLuNYsPcXJ1zhobMVG")
	require.NoError(t, err)

	expBanksPublicKey, err := ParsePublicKey("GzjFRde8rxJrjoCKxDNtnmkxUZKPAuoLmgkqEgjJeVDe")
	require.NoError(t, err)

	alias, err := ParseIdentityAlias("nimona://id:alias:nimona.dev")
	require.NoError(t, err)

	exp := &IdentityInfo{
		Alias: IdentityAlias{
			Hostname: "nimona.dev",
		},
		Identity: Identity{
			KeyGraph: expKeyGraphID.DocumentHash,
			Use:      "provider",
		},
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

	fmt.Println(got.Identity.String())
	fmt.Println(got.Identity.KeyGraph.String())
	fmt.Println(got.PeerAddresses[0].PublicKey.String())
	fmt.Println(got.PeerAddresses[1].PublicKey.String())

	require.NoError(t, err)
	require.Equal(t, exp, got)
	require.Equal(t, expIdentityShort, got.Identity.String())
}
