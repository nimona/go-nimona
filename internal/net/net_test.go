package net

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

func TestNetConnectionSuccess(t *testing.T) {
	ctx := context.New()

	n1 := newPeer(t)
	n2 := newPeer(t)

	_, err := n1.Listen(ctx, "127.0.0.1:0", &ListenConfig{
		BindLocal: true,
	})
	assert.NoError(t, err)
	time.Sleep(time.Second)

	done := make(chan bool)

	resObj := &object.Object{
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}

	// attempt to dial own address, should fail
	t.Run("dial own address, should error", func(t *testing.T) {
		_, err = n1.Dial(ctx, &peer.ConnectionInfo{
			PublicKey: n1.peerKey.PublicKey(),
			Addresses: n1.Addresses(),
		})
		require.Equal(t, ErrAllAddressesBlocked, err)
	})

	// dial n1 from n2
	t.Run("dial n1 from n2", func(t *testing.T) {
		// wait for new connections on n1
		scs := make(chan Connection)
		n1.RegisterConnectionHandler(func(c Connection) {
			fmt.Println("handing connection on n1")
			scs <- c
		})

		go func() {
			cconn, err := n2.Dial(ctx, &peer.ConnectionInfo{
				PublicKey: n1.peerKey.PublicKey(),
				Addresses: n1.Addresses(),
			})
			assert.NoError(t, err)
			fmt.Println("dialed n2")
			time.Sleep(time.Second)
			fmt.Println("writing foobar to n2")
			err = cconn.Write(context.Background(), resObj)
			assert.NoError(t, err)
			fmt.Println("wrote foobar to n2, done")
			done <- true
		}()

		// wait for connection
		var sc Connection
		select {
		case sc = <-scs:
		case <-time.After(time.Second):
			t.Error("timed out waiting for connection")
			t.FailNow()
		}

		// start listening for incoming messages
		fmt.Println("waiting for messages")
		scReader := sc.Read(ctx)

		// wait for ping message
		fmt.Println("waiting for ping")
		gotObj, err := scReader.Read()
		require.NoError(t, err)
		require.Equal(t, "ping", gotObj.Type)

		// wait for foobar message
		fmt.Println("waiting for foobar")
		gotObj, err = scReader.Read()
		require.NoError(t, err)
		require.EqualValues(t, resObj, gotObj)

		// wait for done
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("timed out waiting for done")
			t.Fail()
		}
	})
}

func TestNetDialBackoff(t *testing.T) {
	s1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	s2, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	ctx := context.New()
	p := &peer.ConnectionInfo{
		PublicKey: s1.PublicKey(),
		Addresses: []string{"tcps:240.0.0.1:1000"},
	}

	p2 := &peer.ConnectionInfo{
		PublicKey: s2.PublicKey(),
		Addresses: p.Addresses,
	}

	// attempt 1, failed
	n1 := newPeer(t)
	_, err = n1.Dial(ctx, p)
	assert.Equal(t, ErrAllAddressesFailed, err)

	// attempt 2, blocked
	_, err = n1.Dial(ctx, p)
	assert.Equal(t, ErrAllAddressesBlocked, err)

	// attempt 2, same address different key, failed
	_, err = n1.Dial(ctx, p2)
	assert.Equal(t, ErrAllAddressesFailed, err)

	// wait for backoff to expire
	time.Sleep(time.Second * 2)

	// attempt 3, failed
	_, err = n1.Dial(ctx, p)
	assert.Equal(t, ErrAllAddressesFailed, err)

	// attempt 4, blocked
	_, err = n1.Dial(ctx, p)
	assert.Equal(t, ErrAllAddressesBlocked, err)

	// wait, but backoff should not have expired
	time.Sleep(time.Second * 2)

	// attempt 5, blocked
	_, err = n1.Dial(ctx, p)
	assert.Equal(t, ErrAllAddressesBlocked, err)

	// wait for backoff to expire
	time.Sleep(time.Second * 2)

	// attempt 6, failed
	_, err = n1.Dial(ctx, p)
	assert.Equal(t, ErrAllAddressesFailed, err)
}

func newPeer(t *testing.T) *network {
	pk, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)
	return New(
		pk,
	).(*network)
}
