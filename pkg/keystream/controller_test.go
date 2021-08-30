package keystream

import (
	"database/sql"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/sqlobjectstore"
)

func TestController_New(t *testing.T) {
	sqlStoreDB, err := sql.Open(
		"sqlite3",
		path.Join(t.TempDir(), "db.sqlite"),
	)
	require.NoError(t, err)
	sqlStore, err := sqlobjectstore.New(sqlStoreDB)
	require.NoError(t, err)

	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	// create a controller with empty stores
	ctrl, err := NewController(k.PublicKey().DID(), sqlStore, sqlStore, nil)
	require.NoError(t, err)
	require.NotNil(t, ctrl)

	// get active and next keys
	wantActiveKey := ctrl.currentPrivateKey.PublicKey()
	wantNextKeyDigest := ctrl.state.NextKeyDigest

	// restore controller from same stores
	ctrl2, err := RestoreController(ctrl.state.Root, sqlStore, sqlStore)
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
	require.Equal(t, wantNextKeyDigest, rotationEvent.Key.Hash())

	gotActiveKey = ctrl2.currentPrivateKey.PublicKey()
	gotNextKeyHash = ctrl2.state.NextKeyDigest

	require.NotEqual(t, wantActiveKey, gotActiveKey)
	require.NotEqual(t, wantNextKeyDigest, gotNextKeyHash)
}
