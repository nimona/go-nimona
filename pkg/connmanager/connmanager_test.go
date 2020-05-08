package connmanager

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/net"
	"nimona.io/pkg/peer"
)

func TestGetConnection(t *testing.T) {
	ctx := context.Background()

	handler := func(conn *net.Connection) error {
		return nil
	}

	eb1, _, n1 := newPeer(t)
	eb2, kc2, n2 := newPeer(t)

	mgr := New(ctx, eb1, n1, handler)

	lst1, err := n1.Listen(ctx, "0.0.0.0:0")
	assert.NoError(t, err)
	defer lst1.Close()

	mgr2 := New(ctx, eb2, n2, handler)
	assert.NotNil(t, mgr2)

	lst2, err := n2.Listen(ctx, "0.0.0.0:0")
	assert.NoError(t, err)
	defer lst2.Close()

	conn1, err := mgr.GetConnection(ctx, &peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
	})
	assert.NoError(t, err)

	conn2, err := mgr.GetConnection(ctx, &peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
	})
	assert.NoError(t, err)

	// verify that we retrieved the same connection
	if !reflect.DeepEqual(conn1, conn2) {
		t.Errorf("manager.GetConnection() = %v, want %v", conn1, conn2)
	}
}

func newPeer(t *testing.T) (
	eventbus.Eventbus,
	keychain.Keychain,
	net.Network,
) {
	net.BindLocal = true

	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	eb := eventbus.New()

	kc := keychain.New()
	kc.Put(keychain.PrimaryPeerKey, pk)

	return eb, kc, net.New(
		net.WithKeychain(kc),
		net.WithEventBus(eb),
	)
}
