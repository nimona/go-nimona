package net

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestNetConnectionSuccess(t *testing.T) {
	ctx := context.New()

	kc1, n1 := newPeer(t)
	_, n2 := newPeer(t)

	_, err := n1.Listen(ctx, "127.0.0.1:0", &ListenConfig{
		BindLocal: true,
	})
	assert.NoError(t, err)

	done := make(chan bool)

	resObj := object.FromMap(object.Map{ // nolint: errcheck
		"foo": object.String("bar"),
	})

	go func() {
		cconn, err := n2.Dial(ctx, &peer.ConnectionInfo{
			PublicKey: kc1.GetPeerKey().PublicKey(),
			Addresses: n1.Addresses(),
		})
		assert.NoError(t, err)
		err = Write(resObj, cconn)
		assert.NoError(t, err)
		done <- true
	}()

	// attempt to dial own address, should fail
	_, err = n1.Dial(ctx, &peer.ConnectionInfo{
		PublicKey: kc1.GetPeerKey().PublicKey(),
		Addresses: n1.Addresses(),
	})
	require.Equal(t, ErrAllAddressesBlocked, err)

	sc, err := n1.Accept()
	require.NoError(t, err)

	reqObj := object.FromMap(object.Map{
		"foo": object.String("bar"),
	})
	err = Write(reqObj, sc)
	assert.NoError(t, err)

	gotObj, err := Read(sc)
	require.NoError(t, err)
	assert.Equal(t, "ping", gotObj.Type)

	gotObj, err = Read(sc)
	require.NoError(t, err)
	assert.EqualValues(t, resObj.ToMap(), gotObj.ToMap())

	<-done
}

func TestNetDialBackoff(t *testing.T) {
	s1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	s2, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
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
	_, n1 := newPeer(t)
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

func newPeer(t *testing.T) (
	localpeer.LocalPeer,
	*network,
) {
	kc := localpeer.New()
	pk, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	assert.NoError(t, err)
	kc.SetPeerKey(pk)
	return kc, New(
		kc,
	).(*network)
}
