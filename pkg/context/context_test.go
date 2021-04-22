package context_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/context"
)

func TestContext(t *testing.T) {
	method := "TestContext"
	args := map[string]interface{}{
		"foo": "bar",
	}
	cid := "001"
	ctx := context.New(
		context.WithMethod(method),
		context.WithArguments(args),
		context.WithCorrelationID(cid),
	)
	assert.Equal(t, method, ctx.Method())
	assert.Equal(t, args, ctx.Arguments())
	assert.Equal(t, cid, ctx.CorrelationID())
}

func TestContextCancel(t *testing.T) {
	ctx := context.New(
		context.WithCancel(),
	)

	go func() {
		time.Sleep(time.Millisecond * 50)
		ctx.Cancel()
	}()

	done := false
	select {
	case <-ctx.Done():
		done = true
	case <-time.After(time.Millisecond * 250):
		t.Errorf("context didn't cancel on time")
	}

	assert.True(t, done)
}

func TestContextNestedCancel(t *testing.T) {
	ctx := context.New(
		context.WithCancel(),
	)
	cctx := context.New(
		context.WithParent(ctx),
		context.WithCancel(),
	)

	go func() {
		time.Sleep(time.Millisecond * 50)
		ctx.Cancel()
	}()

	done := false
	select {
	case <-cctx.Done():
		done = true
	case <-time.After(time.Millisecond * 250):
		t.Errorf("context didn't cancel on time")
	}

	assert.True(t, done)
}

func TestContextTimout(t *testing.T) {
	ctx := context.New(
		context.WithTimeout(time.Millisecond * 150),
	)

	done := false
	select {
	case <-ctx.Done():
		done = true
	case <-time.After(time.Millisecond * 250):
		t.Errorf("context didn't timeout on time")
	}

	assert.True(t, done)
}
