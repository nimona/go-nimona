package stream_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/stream"
	"nimona.io/pkg/tilde"
)

func TestSyncStrategy_Integration(t *testing.T) {
	_, c0 := provider.NewTestProvider(context.Background(), t)

	k0, err := crypto.PublicKeyFromDID(c0.Metadata.Owner)
	require.NoError(t, err)

	d1, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath(t.TempDir()),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultBootstraps([]peer.Shorthand{
				peer.Shorthand(fmt.Sprintf("%s@%s", k0, c0.Addresses[0])),
			}),
		),
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
				peer.Shorthand(fmt.Sprintf("%s@%s", k0, c0.Addresses[0])),
			}),
		),
	)
	require.NoError(t, err)

	m2 := d2.StreamManager()

	time.Sleep(time.Second)

	// directly add the objects to p1's store, without going through the stream
	// manager.
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
	require.NoError(t, err)
	require.Len(t, o1gs, 1)
	require.Equal(t, o1, o1gs[0])

	start := time.Now()
	c2, err := m2.GetOrCreateController(h1)
	require.NoError(t, err)

	time.Sleep(time.Second)

	// attempt to fetch the stream using the stream manager on p2.
	n, err := m2.Fetch(context.New(), c2, h1)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	fmt.Println("---", time.Since(start))
}

func TestSyncStrategy_Announcements_Integration(t *testing.T) {
	_, c0 := provider.NewTestProvider(context.Background(), t)

	k0, err := crypto.PublicKeyFromDID(c0.Metadata.Owner)
	require.NoError(t, err)

	d1, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath(t.TempDir()),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultBootstraps([]peer.Shorthand{
				peer.Shorthand(fmt.Sprintf("%s@%s", k0, c0.Addresses[0])),
			}),
		),
	)
	require.NoError(t, err)

	d2, err := daemon.New(
		context.New(),
		daemon.WithConfigOptions(
			config.WithDefaultPath(t.TempDir()),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultBootstraps([]peer.Shorthand{
				peer.Shorthand(fmt.Sprintf("%s@%s", k0, c0.Addresses[0])),
			}),
		),
	)
	require.NoError(t, err)

	m1 := d1.StreamManager()
	m2 := d2.StreamManager()

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
