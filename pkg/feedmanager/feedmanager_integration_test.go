package feedmanager_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/fixtures"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/keystream"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/testutils"
)

func Test_Manager_Integration(t *testing.T) {
	// construct local bootstrap peer
	bootstrapConnectionInfo := testutils.NewTestBootstrapPeer(t)

	time.Sleep(time.Second * 2)

	// construct new identity key
	id, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	// construct peer 0
	p0 := newDaemon(t, "p0", id, bootstrapConnectionInfo)
	_, err = p0.KeyStreamManager().NewController(nil)
	require.NoError(t, err)
	defer p0.Close()
	time.Sleep(time.Second * 2)

	// construct peer 1
	p1 := newDaemon(t, "p1", id, bootstrapConnectionInfo)
	defer p1.Close()
	time.Sleep(time.Second * 2)

	// create new delegation request
	dr1, c1ch, err := p1.KeyStreamManager().NewDelegationRequest(
		context.Background(), // no timeout
		keystream.DelegationRequestVendor{},
		keystream.Permissions{},
	)
	require.NoError(t, err)

	// pass dr to delegator handler
	err = p0.KeyStreamManager().HandleDelegationRequest(
		context.Background(),
		dr1,
	)
	require.NoError(t, err)

	// and wait for the controller
	c1 := <-c1ch
	require.NotNil(t, c1)

	// wait for resolver to pick up the new controller and publish a new
	// connection info
	time.Sleep(time.Second * 2)

	// put a new stream on p0
	o0 := object.MustMarshal(
		&fixtures.TestStream{
			Metadata: object.Metadata{
				Owner: id.PublicKey().DID(),
			},
			Nonce: "foo",
		},
	)
	err = p0.ObjectManager().Put(context.TODO(), o0)
	require.NoError(t, err)

	time.Sleep(time.Second * 5)

	// wait until resolver sees provider
	found := false
	for i := 0; i < 10; i++ {
		r, err := p1.Resolver().Lookup(
			context.New(
				context.WithCorrelationID("p1.resolver.lookup"),
				context.WithTimeout(time.Second),
			),
			resolver.LookupByDigest(o0.Hash()),
		)
		if err != nil {
			continue
		}
		if len(r) > 0 {
			found = true
			break
		}
		time.Sleep(time.Second * 2)
	}
	require.True(t, found)

	time.Sleep(time.Second * 5)

	fmt.Println(">>> p0 pk", p0.Network().GetConnectionInfo().PublicKey.String())
	fmt.Println(">>> p1 pk", p1.Network().GetConnectionInfo().PublicKey.String())

	c0, err := p0.KeyStreamManager().GetController()
	require.NoError(t, err)
	fmt.Println(">>> p0 id", c0.GetKeyStream().GetDID())
	c1, err = p1.KeyStreamManager().GetController()
	require.NoError(t, err)
	fmt.Println(">>> p1 id", c1.GetKeyStream().GetDID())

	// wait a bit, and check stream on p1
	fmt.Println("????", o0.Hash().String())
	g0, err := p1.ObjectStore().Get(o0.Hash())
	require.NoError(t, err)
	require.NotNil(t, g0)
}

func newDaemon(
	t *testing.T,
	name string,
	id crypto.PrivateKey,
	bootstrapConnectionInfo *peer.ConnectionInfo,
) daemon.Daemon {
	tempDir := t.TempDir()
	fmt.Println("tempDir", tempDir)
	d, err := daemon.New(
		context.New(
			context.WithCorrelationID(name),
		),
		daemon.WithConfigOptions(
			config.WithDefaultPath(tempDir),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultBootstraps([]peer.Shorthand{
				peer.Shorthand(
					fmt.Sprintf(
						"%s@%s",
						bootstrapConnectionInfo.PublicKey.String(),
						bootstrapConnectionInfo.Addresses[0],
					),
				),
			},
			),
		),
	)
	require.NoError(t, err)

	d.FeedManager().RegisterFeed(
		fixtures.TestStreamType,
	)

	time.Sleep(time.Second)
	return d
}
