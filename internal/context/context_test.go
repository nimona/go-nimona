package context_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/internal/context"
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
