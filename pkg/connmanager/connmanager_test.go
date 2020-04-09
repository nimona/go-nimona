package connmanager

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

func TestGetConnection(t *testing.T) {
	ctx := context.Background()

	n, li, disc := newPeer(t)
	n2, li2, _ := newPeer(t)
	disc.Add(li2.GetSignedPeer(), true)

	mgr, err := New(ctx, n, li)
	assert.NoError(t, err)

	mgr2, err := New(ctx, n2, li2)
	assert.NoError(t, err)
	assert.NotNil(t, mgr2)

	conn1, err := mgr.GetConnection(ctx, li2.GetSignedPeer())
	conn2, err := mgr.GetConnection(ctx, li2.GetSignedPeer())
	assert.NoError(t, err)

	// assert that we retrived the same connection
	if !reflect.DeepEqual(conn1, conn2) {
		t.Errorf("manager.GetConnection() = %v, want %v", conn1, conn2)
	}
}

func newPeer(t *testing.T) (
	net.Network,
	*peer.LocalPeer,
	discovery.PeerStorer,
) {
	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	store, err := sqlobjectstore.New(tempSqlite3(t))
	assert.NoError(t, err)

	discover := discovery.NewPeerStorer(store)

	li, err := peer.NewLocalPeer("", pk)
	assert.NoError(t, err)

	n, err := net.New(discover, li)
	assert.NoError(t, err)

	tcp := net.NewTCPTransport(li, "0.0.0.0:0")
	n.AddTransport("tcps", tcp)

	hsm := handshake.New(li, discover)
	n.AddMiddleware(hsm.Handle())

	// _, err = exchange.New(context.Background(), pk, n, store, discover, li)

	return n, li, discover
}

func tempSqlite3(t *testing.T) *sql.DB {
	dirPath, err := ioutil.TempDir("", "nimona-store-sql")
	require.NoError(t, err)
	fmt.Println(path.Join(dirPath, "sqlite3.db"))
	db, err := sql.Open("sqlite3", path.Join(dirPath, "sqlite3.db"))
	require.NoError(t, err)
	return db
}
