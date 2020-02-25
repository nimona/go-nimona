package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoadFromFile(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	c := New()
	c.Path = "./"

	err := c.Load()
	assert.NoError(t, err)
	assert.Equal(t, "./", c.Path)

	assert.Equal(t, "ed25519.prv.foo", c.Peer.IdentityKey.String())
	assert.Equal(t, "ed25519.prv.foo", c.Peer.PeerKey.String())
	assert.Equal(t, []string{"tcps:foo:21013"}, c.Peer.BootstrapAddresses)
}

func TestConfigLoadFromEnvs(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	c := New()
	c.Path = "./missing"

	os.Setenv("NIMONA_PEER_IDENTITY_KEY", "id1")            // nolint: errcheck
	os.Setenv("NIMONA_PEER_KEY", "pk1")                     // nolint: errcheck
	os.Setenv("NIMONA_PEER_HOSTNAME", "foo")                // nolint: errcheck
	os.Setenv("NIMONA_PEER_BOOTSTRAP_ADDRESSES", "bs1,bs2") // nolint: errcheck

	err := c.Load()
	assert.NoError(t, err)
	assert.Equal(t, "./missing", c.Path)

	assert.Equal(t, "foo", c.Peer.AnnounceHostname)
	assert.Equal(t, "id1", c.Peer.IdentityKey.String())
	assert.Equal(t, "pk1", c.Peer.PeerKey.String())
	assert.Equal(t, []string{"bs1", "bs2"}, c.Peer.BootstrapAddresses)
}

func TestConfigLoadFromFileAndEnvs(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	c := New()
	c.Path = "./"

	os.Setenv("NIMONA_PEER_IDENTITY_KEY", "id1")            // nolint: errcheck
	os.Setenv("NIMONA_PEER_HOSTNAME", "foo")                // nolint: errcheck
	os.Setenv("NIMONA_PEER_BOOTSTRAP_ADDRESSES", "bs1,bs2") // nolint: errcheck

	err := c.Load()
	assert.NoError(t, err)
	assert.Equal(t, "./", c.Path)

	assert.Equal(t, "foo", c.Peer.AnnounceHostname)
	assert.Equal(t, "id1", c.Peer.IdentityKey.String())
	assert.Equal(t, "ed25519.prv.foo", c.Peer.PeerKey.String())
	assert.Equal(t, []string{"bs1", "bs2"}, c.Peer.BootstrapAddresses)
}

func TestConfigLoadFromFileAndEnvsEmpty(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	c := New()
	c.Path = "./"

	os.Setenv("NIMONA_PEER_IDENTITY_KEY", "id1")     // nolint: errcheck
	os.Setenv("NIMONA_PEER_HOSTNAME", "foo")         // nolint: errcheck
	os.Setenv("NIMONA_PEER_BOOTSTRAP_ADDRESSES", "") // nolint: errcheck

	err := c.Load()
	assert.NoError(t, err)
	assert.Equal(t, "./", c.Path)

	assert.Equal(t, "foo", c.Peer.AnnounceHostname)
	assert.Equal(t, "id1", c.Peer.IdentityKey.String())
	assert.Equal(t, "ed25519.prv.foo", c.Peer.PeerKey.String())
	assert.Equal(t, []string{}, c.Peer.BootstrapAddresses)
}

func TestConfigLoadFromFileAndEnvsMissing(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	c := New()
	c.Path = "./"

	os.Setenv("NIMONA_PEER_IDENTITY_KEY", "id1") // nolint: errcheck
	os.Setenv("NIMONA_PEER_HOSTNAME", "foo")     // nolint: errcheck

	err := c.Load()
	assert.NoError(t, err)
	assert.Equal(t, "./", c.Path)

	assert.Equal(t, "foo", c.Peer.AnnounceHostname)
	assert.Equal(t, "id1", c.Peer.IdentityKey.String())
	assert.Equal(t, "ed25519.prv.foo", c.Peer.PeerKey.String())
	assert.Equal(t, []string{"tcps:foo:21013"}, c.Peer.BootstrapAddresses)
}
