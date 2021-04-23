package context

import (
	stdcontext "context"
	"time"

	"nimona.io/internal/rand"
)

type (
	// Context that matches std context
	Context interface {
		Cancel()
		Deadline() (deadline time.Time, ok bool)
		Done() <-chan struct{}
		Err() error
		Value(key interface{}) interface{}
		Arguments() map[string]interface{}
		Method() string
		CorrelationID() string
	}
	// context wraps stdcontext.Context allowing adding tracing information
	// instead of using the Values.
	context struct {
		stdcontext.Context
		method        string
		arguments     map[string]interface{}
		correlationID string
		timeout       time.Duration
		withCancel    bool
		cancel        func()
	}
)

var (
	todo = New()
)

// TODO returns a non-nil, empty Context. Code should use context.TODO when
// it's unclear which Context to use or it is not yet available.
func TODO() Context {
	return todo
}

// Background context wrapper
func Background() Context {
	return New()
}

// A CancelFunc tells an operation to abandon its work
type CancelFunc func()

// Method returns the context's method
func (ctx *context) Method() string {
	return ctx.method
}

// Cancel the context
func (ctx *context) Cancel() {
	if ctx.cancel != nil {
		ctx.cancel()
	}
}

// Arguments returns the context's arguments
func (ctx *context) Arguments() map[string]interface{} {
	return ctx.arguments
}

// CorrelationID returns the context's correlationID
func (ctx *context) CorrelationID() string {
	if ctx.correlationID != "" {
		return ctx.correlationID
	}

	if ctx.Context != nil {
		return GetCorrelationID(ctx.Context)
	}

	return ""
}

// FromContext returns a new context from a parent
func FromContext(ctx stdcontext.Context) Context {
	return New(WithParent(ctx))
}

// GetCorrelationID returns the correlation if there is one
func GetCorrelationID(ctx stdcontext.Context) string {
	switch cctx := ctx.(type) {
	case *context:
		return cctx.CorrelationID()
	default:
		return ""
	}
}

// New constructs a new *context from a parent Context and Options
func New(opts ...Option) Context {
	ctx := &context{
		Context:   stdcontext.Background(),
		arguments: map[string]interface{}{},
	}
	for _, opt := range opts {
		opt(ctx)
	}
	if ctx.correlationID == "" {
		ctx.correlationID = rand.String(8)
	}
	if pctx, ok := ctx.Context.(*context); ok {
		ctx.correlationID = pctx.correlationID + "/" + ctx.correlationID
	}
	if ctx.timeout > 0 {
		ctx.Context, ctx.cancel = stdcontext.WithTimeout(
			ctx.Context,
			ctx.timeout,
		)
	} else if ctx.withCancel {
		ctx.Context, ctx.cancel = stdcontext.WithCancel(
			ctx.Context,
		)
	}
	return ctx
}
