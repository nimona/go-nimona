package keystream

import (
	"database/sql"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xujiajun/nutsdb"

	"nimona.io/pkg/sqlobjectstore"
)

func TestController_New(t *testing.T) {
	opt := nutsdb.DefaultOptions
	opt.Dir = t.TempDir()
	kvStore, err := nutsdb.Open(opt)
	require.NoError(t, err)
	defer kvStore.Close()

	objectStoreDB, err := sql.Open(
		"sqlite3",
		path.Join(t.TempDir(), "db.sqlite"),
	)
	require.NoError(t, err)
	objectStore, err := sqlobjectstore.New(objectStoreDB)
	require.NoError(t, err)

	ctrl, err := NewController(kvStore, objectStore)
	require.NoError(t, err)
	require.NotNil(t, ctrl)

	wantActiveKey := ctrl.currentPrivateKey.PublicKey()
	wantNextKeyDigest := ctrl.state.NextKeyDigest

	ctrl2, err := NewController(kvStore, objectStore)
	require.NoError(t, err)
	require.NotNil(t, ctrl2)

	gotActiveKey := ctrl2.currentPrivateKey.PublicKey()
	gotNextKeyHash := ctrl2.state.NextKeyDigest

	require.NotEmpty(t, gotActiveKey)
	require.NotEmpty(t, gotNextKeyHash)

	require.Equal(t, wantActiveKey, gotActiveKey)
	require.Equal(t, wantNextKeyDigest, gotNextKeyHash)

	rotationEvent, err := ctrl2.Rotate()
	require.NoError(t, err)
	require.NotNil(t, rotationEvent)

	require.Equal(t, wantNextKeyDigest, getPublicKeyHash(rotationEvent.Key))

	gotActiveKey = ctrl2.currentPrivateKey.PublicKey()
	gotNextKeyHash = ctrl2.state.NextKeyDigest

	require.NotEqual(t, wantActiveKey, gotActiveKey)
	require.NotEqual(t, wantNextKeyDigest, gotNextKeyHash)
}
