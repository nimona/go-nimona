package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestHandlerNetwork(t *testing.T) {
	srv, clt := newTestSessionManager(t)

	info := NetworkInfo{
		NetworkAlias: NetworkAlias{
			Hostname: "testing.nimona.io",
		},
		PeerAddresses: []PeerAddr{{
			Network: "utp",
			Address: "localhost:1234",
		}},
	}
	hnd := &HandlerNetwork{
		Hostname: "testing.nimona.io",
		PeerAddresses: []PeerAddr{{
			Network: "utp",
			Address: "localhost:1234",
		}},
	}
	srv.RegisterHandler(
		"core/network/info.request",
		hnd.HandleNetworkInfoRequest,
	)

	// dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// ask for capabilities
	ctx := context.Background()
	res, err := RequestNetworkInfo(ctx, ses)
	require.NoError(t, err)
	require.Equal(t, info.NetworkAlias, res.NetworkAlias)
	require.Equal(t, info.PeerAddresses, res.PeerAddresses)
}

func TestRequestNetworkJoin(t *testing.T) {
	ctx := context.Background()

	// Create new peer configs
	srvPeerConfig := NewTestPeerConfig(t)
	clnPeerConfig := NewTestPeerConfig(t)

	// Create new session manager
	srv, clt := newTestSessionManager(t)

	// Construct a new HandlerNetwork
	hnd := &HandlerNetwork{
		Hostname: "testing.nimona.io",
		PeerAddresses: []PeerAddr{{
			Network: "utp",
			Address: "localhost:1234",
		}},
		PeerConfig: srvPeerConfig,
		Store:      NewTestNetworkStore(t),
	}
	srv.RegisterHandler(
		"core/network/join.request",
		hnd.HandleNetworkJoinRequest,
	)

	// Dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// Test with missing identity
	t.Run("missing identity", func(t *testing.T) {
		_, err = RequestNetworkJoin(ctx, ses, clnPeerConfig, "test-handle")
		require.Error(t, err)
	})

	// Create a new identity for the client
	clnPeerConfig.SetIdentity(NewTestIdentity(t))

	// Test with empty handle
	t.Run("empty handle", func(t *testing.T) {
		res, err := RequestNetworkJoin(ctx, ses, clnPeerConfig, "")
		require.NoError(t, err)
		require.False(t, res.Accepted)
		require.NotEmpty(t, res.Error)
	})

	// Test with a valid identity and non existing handle
	t.Run("valid identity and non existing handle", func(t *testing.T) {
		res, err := RequestNetworkJoin(ctx, ses, clnPeerConfig, "test-handle")
		require.NoError(t, err)
		require.True(t, res.Accepted)
		require.Equal(t, "test-handle", res.Handle)
	})

	// Test with an existing identity
	t.Run("existing identity", func(t *testing.T) {
		res, err := RequestNetworkJoin(ctx, ses, clnPeerConfig, "new-handle")
		require.NoError(t, err)
		require.False(t, res.Accepted)
		require.NotEmpty(t, res.Error)
	})
}

func TestHandlerNetwork_ResolveHandle(t *testing.T) {
	ctx := context.Background()

	// Create new peer configs
	srvPeerConfig := NewTestPeerConfig(t)
	clnPeerConfig := NewTestPeerConfig(t)

	// Create new session manager
	srv, clt := newTestSessionManager(t)

	// Construct a new HandlerNetwork
	hnd := &HandlerNetwork{
		Hostname: "testing.nimona.io",
		PeerAddresses: []PeerAddr{{
			Network: "utp",
			Address: "localhost:1234",
		}},
		PeerConfig: srvPeerConfig,
		Store:      NewTestNetworkStore(t),
	}
	srv.RegisterHandler(
		"core/network/resolveHandle.request",
		hnd.HandleNetworkResolveHandleRequest,
	)
	srv.RegisterHandler(
		"core/network/join.request",
		hnd.HandleNetworkJoinRequest,
	)

	// Dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// Test with empty handle
	t.Run("empty handle", func(t *testing.T) {
		_, err := RequestNetworkResolveHandle(ctx, ses, "")
		require.Error(t, err)
	})

	// Test with non existing handle
	t.Run("non existing handle", func(t *testing.T) {
		res, err := RequestNetworkResolveHandle(ctx, ses, "test-handle")
		require.NoError(t, err)
		require.NotEmpty(t, res.ErrorDescription)
		require.True(t, res.Error)
		require.False(t, res.Found)
	})

	// Create a new identity for the client
	clnPeerConfig.SetIdentity(NewTestIdentity(t))

	// Join with a valid identity and non existing handle
	t.Run("valid identity and non existing handle", func(t *testing.T) {
		res, err := RequestNetworkJoin(ctx, ses, clnPeerConfig, "test-handle")
		require.NoError(t, err)
		require.Empty(t, res.ErrorDescription)
		require.False(t, res.Error)
		require.True(t, res.Accepted)
		require.Equal(t, "test-handle", res.Handle)
	})

	// Test with an existing handle
	t.Run("existing handle", func(t *testing.T) {
		res, err := RequestNetworkResolveHandle(ctx, ses, "test-handle")
		require.NoError(t, err)
		require.True(t, res.Found)
		require.EqualValues(t, clnPeerConfig.GetIdentity(), &res.IdentityID)
	})
}

func TestHandlerNetwork_AnnouncePeer_LookupPeer(t *testing.T) {
	ctx := context.Background()

	// Create new peer configs
	srvPeerConfig := NewTestPeerConfig(t)
	clnPeerConfig := NewTestPeerConfig(t)

	// Create new session manager
	srv, clt := newTestSessionManager(t)

	// Construct a new HandlerNetwork
	hnd := &HandlerNetwork{
		Hostname: "testing.nimona.io",
		PeerAddresses: []PeerAddr{{
			Network: "utp",
			Address: "localhost:1234",
		}},
		PeerConfig: srvPeerConfig,
		Store:      NewTestNetworkStore(t),
	}
	srv.RegisterHandler(
		"core/network/announcePeer.request",
		hnd.HandleNetworkAnnouncePeerRequest,
	)
	srv.RegisterHandler(
		"core/network/lookupPeer.request",
		hnd.HandleNetworkLookupPeerRequest,
	)

	// Dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// Create a new identity for the client
	clnPeerConfig.SetIdentity(NewTestIdentity(t))

	// Announce the client
	t.Run("announce client's peer", func(t *testing.T) {
		res, err := RequestNetworkAnnouncePeer(ctx, ses, clnPeerConfig)
		require.NoError(t, err)
		require.Empty(t, res.ErrorDescription)
		require.False(t, res.Error)
	})

	// Lookup the client
	t.Run("lookup client", func(t *testing.T) {
		res, err := RequestNetworkLookupPeer(ctx, ses, *clnPeerConfig.GetPeerKey())
		require.NoError(t, err)
		require.Empty(t, res.ErrorDescription)
		require.False(t, res.Error)
		require.True(t, res.Found)
		require.Equal(t, *clnPeerConfig.GetPeerInfo(), res.PeerInfo)
	})
}

func NewTestNetworkStore(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&NetworkAccountingModel{},
		&NetworkPeerModel{},
	)
	require.NoError(t, err)

	return db
}
