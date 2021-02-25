package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/config"
)

func TestConfig(t *testing.T) {
	configPath := t.TempDir()

	type ExtraCfg struct {
		Hello string
	}

	fmt.Println(configPath)

	h1, err := config.New(
		config.WithDefaultPath(configPath),
		config.WithDefaultFilename("nim.json"),
		config.WithExtraConfig("extraOne", &ExtraCfg{
			Hello: "one",
		}),
		config.WithExtraConfig("EXTRA_TWO", &ExtraCfg{
			Hello: "two",
		}),
	)
	assert.NotNil(t, h1)
	assert.NoError(t, err)

	extraTwo := &ExtraCfg{}
	h2, err := config.New(
		config.WithDefaultPath(configPath),
		config.WithDefaultFilename("nim.json"),
		config.WithExtraConfig("extraTwo", extraTwo),
	)
	assert.NotNil(t, h2)
	assert.NoError(t, err)

	assert.Equal(t, h1.Peer.PrivateKey, h2.Peer.PrivateKey)
	assert.Equal(t, "two", extraTwo.Hello)
}

func TestConfigUnmarshal(t *testing.T) {
	type ExtraCfg struct {
		Hello string
	}

	extraConfig1 := &ExtraCfg{}
	extraConfig2 := &ExtraCfg{}

	h1, err := config.New(
		config.WithDefaultPath("."),
		config.WithDefaultFilename("test_config.json"),
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
		config.WithDefaultPath("."),
		config.WithDefaultFilename("test_config.json"),
		config.WithExtraConfig("extraOne", extraConfig1),
		config.WithExtraConfig("extraTwo", extraConfig2),
	)
	assert.NotNil(t, h1)
	assert.NoError(t, err)
	assert.Equal(t, "envar", extraConfig1.Hello)
	assert.Equal(t, "two", extraConfig2.Hello)
}
