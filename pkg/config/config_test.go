package config_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/config"
)

func TestConfig(t *testing.T) {
	tempRootDir := os.TempDir()
	tempDir, err := ioutil.TempDir(tempRootDir, "nimona-test")
	require.NoError(t, err)
	configPath := filepath.Join(tempDir, "nimona")

	type ExtraCfg struct {
		Hello string
	}

	h1, err := config.New(
		config.WithPath(configPath),
		config.WithFilename("nim.json"),
		config.WithExtraConfig("extraOne", &ExtraCfg{
			Hello: "one",
		}),
		config.WithExtraConfig("extraTwo", &ExtraCfg{
			Hello: "two",
		}),
	)
	assert.NotNil(t, h1)
	assert.NoError(t, err)

	extraTwo := &ExtraCfg{}
	h2, err := config.New(
		config.WithPath(configPath),
		config.WithFilename("nim.json"),
		config.WithExtraConfig("extraTwo", extraTwo),
	)
	assert.NotNil(t, h2)
	assert.NoError(t, err)

	assert.Equal(t, h1.Peer.PrivateKey, h2.Peer.PrivateKey)

	assert.NoError(t, err)

	assert.Equal(t, "two", extraTwo.Hello)
}

func TestConfigUnmarshal(t *testing.T) {
	type ExtraCfg struct {
		Hello string
	}

	extraConfig1 := &ExtraCfg{}
	extraConfig2 := &ExtraCfg{}

	h1, err := config.New(
		config.WithPath("."),
		config.WithFilename("test_config.json"),
		config.WithExtraConfig("extraOne", extraConfig1),
		config.WithExtraConfig("extraTwo", extraConfig2),
	)
	assert.NotNil(t, h1)
	assert.NoError(t, err)
	assert.Equal(t, "one", extraConfig1.Hello)
	assert.Equal(t, "two", extraConfig2.Hello)
}

func TestConfigEnvar(t *testing.T) {
	type ExtraCfg struct {
		Hello string `envconfig:"HELLO"`
	}

	os.Setenv("NIMONA_EXTRAONE_HELLO", "envar")

	extraConfig1 := &ExtraCfg{}
	extraConfig2 := &ExtraCfg{}

	h1, err := config.New(
		config.WithPath("."),
		config.WithFilename("test_config.json"),
		config.WithExtraConfig("extraOne", extraConfig1),
		config.WithExtraConfig("extraTwo", extraConfig2),
	)
	assert.NotNil(t, h1)
	assert.NoError(t, err)
	assert.Equal(t, "envar", extraConfig1.Hello)
	assert.Equal(t, "two", extraConfig2.Hello)
}
