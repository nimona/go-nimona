package connmanager

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/peer"
)

func TestGetConnection(t *testing.T) {
	ctx := context.Background()

	handler := func(conn *net.Connection) error {
		return nil
	}

	_, n1 := newPeer(t)
	kc2, n2 := newPeer(t)

	mgr := New(ctx, n1, handler)

	lst1, err := n1.Listen(ctx, "127.0.0.1:0", &net.ListenConfig{
		BindLocal: true,
	})
	assert.NoError(t, err)
	defer lst1.Close()

	mgr2 := New(ctx, n2, handler)
	assert.NotNil(t, mgr2)

	lst2, err := n2.Listen(ctx, "127.0.0.1:0", &net.ListenConfig{
		BindLocal: true,
	})
	assert.NoError(t, err)
	defer lst2.Close()

	conn1, err := mgr.GetConnection(ctx, &peer.ConnectionInfo{
		PublicKey: kc2.GetPrimaryPeerKey().PublicKey(),
		Addresses: n2.Addresses(),
	})
	assert.NoError(t, err)

	conn2, err := mgr.GetConnection(ctx, &peer.ConnectionInfo{
		PublicKey: kc2.GetPrimaryPeerKey().PublicKey(),
		Addresses: n2.Addresses(),
	})
	assert.NoError(t, err)

	// verify that we retrieved the same connection
	if !reflect.DeepEqual(conn1, conn2) {
		t.Errorf("manager.GetConnection() = %v, want %v", conn1, conn2)
	}
}

func newPeer(t *testing.T) (
	localpeer.LocalPeer,
	net.Network,
) {
	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	kc := localpeer.New()
	kc.PutPrimaryPeerKey(pk)

	return kc, net.New(
		kc,
	)
}
