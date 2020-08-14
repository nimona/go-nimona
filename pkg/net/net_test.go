package net

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestNetConnectionSuccess(t *testing.T) {
	ctx := context.New()

	BindLocal = true
	kc1, n1 := newPeer(t)
	_, n2 := newPeer(t)

	_, err := n1.Listen(ctx, "0.0.0.0:0")
	assert.NoError(t, err)

	done := make(chan bool)

	resObj := object.FromMap(map[string]interface{}{ // nolint: errcheck
		"data:m": map[string]interface{}{
			"foo:s": "bar",
		},
	})

	go func() {
		cconn, err := n2.Dial(ctx, &peer.Peer{
			Metadata: object.Metadata{
				Owner: kc1.GetPrimaryPeerKey().PublicKey(),
			},
			Addresses: n1.Addresses(),
		})
		assert.NoError(t, err)
		err = Write(resObj, cconn)
		assert.NoError(t, err)
		done <- true
	}()

	// attempt to dial own address, should fail
	_, err = n1.Dial(ctx, &peer.Peer{
		Metadata: object.Metadata{
			Owner: kc1.GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: n1.Addresses(),
	})
	require.Equal(t, ErrAllAddressesBlocked, err)

	sc, err := n1.Accept()
	require.NoError(t, err)

	reqObj := object.FromMap(map[string]interface{}{ // nolint: errcheck
		"data:m": map[string]interface{}{
			"foo:s": "bar",
		},
	})
	err = Write(reqObj, sc)
	assert.NoError(t, err)

	gotObj, err := Read(sc)
	require.NoError(t, err)
	assert.EqualValues(t, resObj.ToMap(), gotObj.ToMap())

	<-done
}

func TestNetDialBackoff(t *testing.T) {
	ctx := context.New()
	p := &peer.Peer{
		Metadata: object.Metadata{
			Owner: crypto.PublicKey("foo"),
		},
		Addresses: []string{"tcps:240.0.0.1:1000"},
	}

	// attempt 1, failed
	_, n1 := newPeer(t)
	_, err := n1.Dial(ctx, p)
	assert.Equal(t, ErrAllAddressesFailed, err)

	// attempt 2, blocked
	_, err = n1.Dial(ctx, p)
	assert.Equal(t, ErrAllAddressesBlocked, err)

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
	keychain.Keychain,
	*network,
) {
	kc := keychain.New()
	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)
	kc.Put(keychain.PrimaryPeerKey, pk)
	return kc, New(
		WithKeychain(kc),
	).(*network)
}
