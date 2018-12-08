package net

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/crypto"
	"nimona.io/go/storage"
)

func TestNetResolver(t *testing.T) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("DEBUG_BLOCKS", "true")
	os.Setenv("BIND_LOCAL", "true")
	os.Setenv("UPNP", "false")

	n1, _ := newPeer(t)
	n2, _ := newPeer(t)

	n1.Resolver().Add(n2.GetPeerInfo())
	n2.Resolver().Add(n1.GetPeerInfo())

	p2, err := n1.Resolver().Resolve(n2.key.GetPublicKey().HashBase58())
	assert.NoError(t, err)
	// assert.Equal(t, n2.key.GetPublicKey(), p2.SignerKey)
	assert.Equal(t, n2.key.GetPublicKey().HashBase58(), p2.SignerKey.GetPublicKey().HashBase58())

	p1, err := n2.Resolver().Resolve(n1.key.GetPublicKey().HashBase58())
	assert.NoError(t, err)
	// assert.Equal(t, n1.key.GetPublicKey(), p1.SignerKey)
	assert.Equal(t, n1.key.GetPublicKey().HashBase58(), p1.SignerKey.HashBase58())
}

func newPeer(t *testing.T) (*network, *exchange) {
	tp, err := ioutil.TempDir("", "nimona-test-net-exchange")
	assert.NoError(t, err)

	kp := filepath.Join(tp, "key.cbor")
	sp := filepath.Join(tp, "objects")

	pk, err := crypto.LoadKey(kp)
	assert.NoError(t, err)

	ds := storage.NewDiskStorage(sp)

	n, err := NewNetwork(pk, "")
	assert.NoError(t, err)

	x, err := NewExchange(pk, n, ds, fmt.Sprintf("127.0.0.1:%d", 0))
	assert.NoError(t, err)

	return n.(*network), x.(*exchange)
}
