package net

import (
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

	done := make(chan bool)

	resObj := &object.Object{
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}

	go func() {
		cconn, err := n2.Dial(ctx, &peer.ConnectionInfo{
			PublicKey: n1.peerKey.PublicKey(),
			Addresses: n1.Addresses(),
		})
		assert.NoError(t, err)
		err = Write(resObj, cconn)
		assert.NoError(t, err)
		done <- true
	}()

	// attempt to dial own address, should fail
	_, err = n1.Dial(ctx, &peer.ConnectionInfo{
		PublicKey: n1.peerKey.PublicKey(),
		Addresses: n1.Addresses(),
	})
	require.Equal(t, ErrAllAddressesBlocked, err)

	sc, err := n1.Accept()
	require.NoError(t, err)

	reqObj := &object.Object{
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	err = Write(reqObj, sc)
	assert.NoError(t, err)

	gotObj, err := Read(sc)
	require.NoError(t, err)
	assert.Equal(t, "ping", gotObj.Type)

	gotObj, err = Read(sc)
	require.NoError(t, err)
	assert.EqualValues(t, resObj, gotObj)

	<-done
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
