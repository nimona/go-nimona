package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/config"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

type ExtraCfg struct {
	Hello string
}

func Test(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	h1, err := config.New(
		config.WithDefaultPath("path"),
		config.WithDefaultListenOnLocalIPs(),
		config.WithDefaultListenOnPrivateIPs(),
		config.WithDefaultListenOnExternalPort(),
		config.WithDefaultDefaultPeerBindAddress("addr"),
		config.WithDefaultBootstraps([]peer.Shorthand{"foo", "bar"}),
		config.WithDefaultPrivateKey(k),
		config.WithDefaultListenOnExternalPort(),
		config.WithExtraConfig("one", &ExtraCfg{
			Hello: "one",
		}),
		config.WithExtraConfig("TWO", &ExtraCfg{
			Hello: "two",
		}),
	)
	require.NoError(t, err)

	m1, err := config.ToEnvPairs("NIMONA", h1)
	assert.NotNil(t, h1)
	require.NoError(t, err)

	gm1 := map[string]string{
		"NIMONA_EXTRAS_ONE_HELLO":          "one",
		"NIMONA_EXTRAS_TWO_HELLO":          "two",
		"NIMONA_LOG_LEVEL":                 "DEBUG",
		"NIMONA_PATH":                      "path",
		"NIMONA_PEER_BIND_ADDRESS":         "addr",
		"NIMONA_PEER_BOOTSTRAPS":           "foo,bar",
		"NIMONA_PEER_LISTEN_EXTERNAL_PORT": "true",
		"NIMONA_PEER_LISTEN_LOCAL":         "true",
		"NIMONA_PEER_LISTEN_PRIVATE":       "true",
		"NIMONA_PEER_PRIVATE_KEY":          k.String(),
	}
	assert.Equal(t, gm1, m1)

	h2, err := config.New(
		config.WithAdditionalEnvVars(m1),
		config.WithExtraConfig("one", &ExtraCfg{}),
		config.WithExtraConfig("TWO", &ExtraCfg{}),
	)
	require.NoError(t, err)
	assert.Equal(t, h1, h2)
}
