package context

import (
	stdcontext "context"
	"time"
)

type (
	// Context that matches std context
	Context stdcontext.Context
	// context wraps stdcontext.Context allowing adding tracing information
	// instead of using the Values.
	context struct {
		stdcontext.Context
		method        string
		arguments     map[string]interface{}
		correlationID string
	}
)

// Background context wrapper
func Background() *context {
	return New(stdcontext.Background())
}

// A CancelFunc tells an operation to abandon its work
type CancelFunc func()

// WithCancel returns a copy of parent with a new Done channel
func WithCancel(ctx stdcontext.Context) (*context, CancelFunc) {
	cctx, cf := stdcontext.WithCancel(ctx)
	return New(cctx), CancelFunc(cf)
}

// Method returns the context's method
func (ctx *context) Method() string {
	return ctx.method
}

// WithTimeout wraps stdcontext.WithTimeout
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	cctx, cf := stdcontext.WithTimeout(parent, timeout)
	return New(cctx), CancelFunc(cf)
}

// Arguments returns the context's arguments
func (ctx *context) Arguments() map[string]interface{} {
	return ctx.arguments
}

// CorrelationID returns the context's correlationID
func (ctx *context) CorrelationID() string {
	return ctx.correlationID
}

// New constructs a new *context from a parent Context and Options
func New(parent stdcontext.Context, opts ...Option) *context {
	ctx := &context{
		Context:   parent,
		arguments: map[string]interface{}{},
	}
	for _, opt := range opts {
		opt(ctx)
	}
	if ctx.correlationID == "" {
		// TODO generate a new one
	}
	return ctx
}
