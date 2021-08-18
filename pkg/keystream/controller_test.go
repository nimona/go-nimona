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

	// create a controller with empty stores
	ctrl, err := NewController(kvStore, objectStore, nil)
	require.NoError(t, err)
	require.NotNil(t, ctrl)

	// get active and next keys
	wantActiveKey := ctrl.currentPrivateKey.PublicKey()
	wantNextKeyDigest := ctrl.state.NextKeyDigest

	// create a new controller with the now not empty stores
	ctrl2, err := NewController(kvStore, objectStore, nil)
	require.NoError(t, err)
	require.NotNil(t, ctrl2)

	// get active and next keys
	gotActiveKey := ctrl2.currentPrivateKey.PublicKey()
	gotNextKeyHash := ctrl2.state.NextKeyDigest

	require.NotEmpty(t, gotActiveKey)
	require.NotEmpty(t, gotNextKeyHash)

	// and make sure they are the same
	require.Equal(t, wantActiveKey, gotActiveKey)
	require.Equal(t, wantNextKeyDigest, gotNextKeyHash)

	// rotate the keys on the latest controller
	rotationEvent, err := ctrl2.Rotate()
	require.NoError(t, err)
	require.NotNil(t, rotationEvent)

	// and check the rotation worked
	require.Equal(t, wantNextKeyDigest, getPublicKeyHash(rotationEvent.Key))

	gotActiveKey = ctrl2.currentPrivateKey.PublicKey()
	gotNextKeyHash = ctrl2.state.NextKeyDigest

	require.NotEqual(t, wantActiveKey, gotActiveKey)
	require.NotEqual(t, wantNextKeyDigest, gotNextKeyHash)
}
