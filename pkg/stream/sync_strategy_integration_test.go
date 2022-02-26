package stream_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
	"nimona.io/pkg/tilde"
)

func TestSyncStrategy_Integration(t *testing.T) {
	_, c0 := provider.NewTestProvider(t, context.Background())

	d1, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath(t.TempDir()),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultBootstraps([]peer.Shorthand{
				peer.Shorthand(fmt.Sprintf("%s@%s", c0.PublicKey, c0.Addresses[0])),
			}),
		),
	)
	require.NoError(t, err)

	m1, err := stream.NewManager(
		context.New(),
		d1.Network(),
		d1.Resolver(),
		d1.ObjectStore().(*sqlobjectstore.Store),
	)
	require.NoError(t, err)

	time.Sleep(time.Second)

	d2, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath(t.TempDir()),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultBootstraps([]peer.Shorthand{
				peer.Shorthand(fmt.Sprintf("%s@%s", c0.PublicKey, c0.Addresses[0])),
			}),
		),
	)
	require.NoError(t, err)

	m2, err := stream.NewManager(
		context.New(),
		d2.Network(),
		d2.Resolver(),
		d2.ObjectStore().(*sqlobjectstore.Store),
	)
	require.NoError(t, err)

	time.Sleep(time.Second)

	o1 := &object.Object{
		Type:     "test",
		Metadata: object.Metadata{},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	h1 := o1.Hash()
	err = d1.ObjectStore().Put(o1)
	require.NoError(t, err)

	o1g, err := d1.ObjectStore().Get(h1)
	require.NoError(t, err)
	require.Equal(t, o1, o1g)

	o1r, err := d1.ObjectStore().GetByStream(h1)
	require.NoError(t, err)
	o1gs, err := object.ReadAll(o1r)
	require.Len(t, o1gs, 1)
	require.Equal(t, o1, o1gs[0])

	f1 := stream.NewTopographicalSyncStrategy(
		d1.Network(),
		d1.Resolver(),
		d1.ObjectStore(),
	)
	go f1.Serve(context.New(), m1)

	start := time.Now()
	c2 := stream.NewController(
		h1,
		d2.Network(),
		d2.ObjectStore().(*sqlobjectstore.Store),
	)
	// f2 := stream.NewTopographicalSyncStrategy(
	// 	d2.Network(),
	// 	d2.Resolver(),
	// 	d2.ObjectStore(),
	// )
	n, err := m2.Fetch(context.New(), c2, h1)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	fmt.Println("---", time.Since(start))
}

func TestSyncStrategy_AnnouncementsIntegration(t *testing.T) {
	_, c0 := provider.NewTestProvider(t, context.Background())

	d1, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath(t.TempDir()),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultBootstraps([]peer.Shorthand{
				peer.Shorthand(fmt.Sprintf("%s@%s", c0.PublicKey, c0.Addresses[0])),
			}),
		),
	)
	require.NoError(t, err)

	m1, err := stream.NewManager(
		context.New(),
		d1.Network(),
		d1.Resolver(),
		d1.ObjectStore().(*sqlobjectstore.Store),
	)
	require.NoError(t, err)

	time.Sleep(time.Second)

	d2, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath(t.TempDir()),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultBootstraps([]peer.Shorthand{
				peer.Shorthand(fmt.Sprintf("%s@%s", c0.PublicKey, c0.Addresses[0])),
			}),
		),
	)
	require.NoError(t, err)

	m2, err := stream.NewManager(
		context.New(),
		d2.Network(),
		d2.Resolver(),
		d2.ObjectStore().(*sqlobjectstore.Store),
	)
	require.NoError(t, err)

	f1 := stream.NewTopographicalSyncStrategy(
		d1.Network(),
		d1.Resolver(),
		d1.ObjectStore(),
	)
	go f1.Serve(context.New(), m1)

	f2 := stream.NewTopographicalSyncStrategy(
		d2.Network(),
		d2.Resolver(),
		d2.ObjectStore(),
	)
	go f2.Serve(context.New(), m2)

	time.Sleep(time.Second)

	o1 := &object.Object{
		Type:     "test",
		Metadata: object.Metadata{},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	h1 := o1.Hash()

	c1, err := m1.GetOrCreateController(h1)
	require.NoError(t, err)
	h1g, err := c1.Insert(o1)
	require.NoError(t, err)
	require.Equal(t, h1, h1g)

	o2 := object.MustMarshal(
		&stream.Subscription{
			Metadata: object.Metadata{
				Owner: d2.Network().GetConnectionInfo().Metadata.Owner,
			},
			RootHashes: []tilde.Digest{
				h1,
			},
		},
	)
	_, err = c1.Insert(o2)
	require.NoError(t, err)

	o3 := &object.Object{
		Type:     "test",
		Metadata: object.Metadata{},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	_, err = c1.Insert(o3)
	require.NoError(t, err)

	gs, err := c1.GetSubscribers()
	require.NoError(t, err)
	require.Len(t, gs, 1)
	require.Equal(t, d2.Network().GetConnectionInfo().Metadata.Owner, gs[0])

	time.Sleep(time.Second * 3)

	c2, err := m2.GetController(h1)
	require.NoError(t, err)

	gs, err = c2.GetSubscribers()
	require.NoError(t, err)
	require.Len(t, gs, 1)
	require.Equal(t, d2.Network().GetConnectionInfo().Metadata.Owner, gs[0])
}
