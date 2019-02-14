package context

import (
	stdcontext "context"
)

type (
	// Context augments the context.Context interface with addition methods
	// to allow to retrieve the extra attributes our wrapper context holds
	Context interface {
		stdcontext.Context
		Method() string
		Arguments() map[string]interface{}
		CorrelationID() string
	}
	// context wraps context.Context allowing adding tracing information instead
	// of using the Values.
	context struct {
		stdcontext.Context
		method        string
		arguments     map[string]interface{}
		correlationID string
	}
)

// Method returns the context's method
func (ctx *context) Method() string {
	return ctx.method
}

// Arguments returns the context's arguments
func (ctx *context) Arguments() map[string]interface{} {
	return ctx.arguments
}

// CorrelationID returns the context's correlationID
func (ctx *context) CorrelationID() string {
	return ctx.correlationID
}

// New constructs a new Context from a parent Context and Options
func New(parent stdcontext.Context, opts ...Option) Context {
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
