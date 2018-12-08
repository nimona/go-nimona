package net

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
)

func TestSendSuccess(t *testing.T) {
	os.Setenv("BIND_LOCAL", "true")
	os.Setenv("UPNP", "false")

	n1, x1 := newPeer(t)
	n2, x2 := newPeer(t)

	n1.Resolver().Add(n2.GetPeerInfo())
	n2.Resolver().Add(n1.GetPeerInfo())

	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := encoding.NewObjectFromMap(em1)

	em2 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo2 := encoding.NewObjectFromMap(em2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1BlockHandled := false
	w2BlockHandled := false

	err := crypto.Sign(eo1, n2.key)
	assert.NoError(t, err)

	_, err = x1.Handle("test/msg", func(o *encoding.Object) error {
		assert.Equal(t, eo1.GetRaw("body"), o.GetRaw("body"))
		assert.NotNil(t, eo1.GetSignerKey())
		assert.NotNil(t, o.GetSignerKey())
		assert.Equal(t, eo1.GetSignerKey(), o.GetSignerKey())
		assert.Equal(t, eo1.GetSignerKey().HashBase58(), o.GetSignerKey().HashBase58())
		assert.NotNil(t, eo1.GetSignature())
		assert.NotNil(t, o.GetSignature())
		assert.Equal(t, eo1.GetSignature(), o.GetSignature())
		assert.Equal(t, eo1.GetSignature().HashBase58(), o.GetSignature().HashBase58())
		w1BlockHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	_, err = x2.Handle("tes**", func(o *encoding.Object) error {
		assert.Equal(t, eo2.GetRaw("body"), o.GetRaw("body"))
		assert.Nil(t, eo2.GetSignature())
		assert.Nil(t, o.GetSignature())
		assert.Nil(t, eo2.GetSignerKey())
		assert.Nil(t, o.GetSignerKey())

		w2BlockHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	ctx := context.Background()

	errS1 := x2.Send(ctx, eo1, "peer:"+n1.key.GetPublicKey().HashBase58())
	assert.NoError(t, errS1)

	time.Sleep(time.Second)

	// TODO should be able to send not signed
	errS2 := x1.Send(ctx, eo2, "peer:"+n2.key.GetPublicKey().HashBase58())
	assert.NoError(t, errS2)

	if errS1 == nil && errS2 == nil {
		wg.Wait()
	}

	assert.True(t, w1BlockHandled)
	assert.True(t, w2BlockHandled)
}
