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

	type lala struct {
		Hello string
	}

	h1, err := config.New(
		config.WithPath(configPath),
		config.WithFilename("nim.json"),
		config.WithExtraConfig("lala", lala{
			Hello: "asd",
		}))
	assert.NotNil(t, h1)
	assert.NoError(t, err)

	h2, err := config.New(
		config.WithPath(configPath),
		config.WithFilename("nim.json"),
	)
	assert.NotNil(t, h2)
	assert.NoError(t, err)

	assert.Equal(t, h1.Peer.PrivateKey, h2.Peer.PrivateKey)

	assert.NoError(t, err)

	val := h2.Extras["lala"].(map[string]interface{})
	assert.Equal(t, "asd", val["Hello"].(string))

}
