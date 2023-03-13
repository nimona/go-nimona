package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverHTTP_ResolveIdentityAlias_E2E(t *testing.T) {
	expIdentityShort := "nimona://id:4Ghn6vjWDZYUQjAN5hEWWGdhocj8oroSLHgZibUNTxbQ"

	expKeyGraphID, err := ParseDocumentID("nimona://doc:4Ghn6vjWDZYUQjAN5hEWWGdhocj8oroSLHgZibUNTxbQ")
	require.NoError(t, err)

	expAsimovPublicKey, err := ParsePublicKey("33f64ioT9yjPdKQ39uVsdUgTriBFMSKFR8EXquRDCmtx")
	require.NoError(t, err)

	expBanksPublicKey, err := ParsePublicKey("E1ZmqKeDEaCpkMviyiEK9mMoxLGtnE6VNawnwe9KhLkw")
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

	// fmt.Println(got.Identity.String())
	// fmt.Println(got.Identity.KeyGraphID.String())
	// fmt.Println(got.PeerAddresses[0].PublicKey.String())
	// fmt.Println(got.PeerAddresses[1].PublicKey.String())

	require.NoError(t, err)
	require.Equal(t, exp, got)
	require.Equal(t, expIdentityShort, got.Identity.String())
}
