package context

// Option allows configuring the context
type Option func(*context)

// WithMethod sets the Context method name
func WithMethod(method string) Option {
	return func(ctx *context) {
		ctx.method = method
	}
}

// WithArguments sets the Context arguments
func WithArguments(args map[string]interface{}) Option {
	// TODO this should loop and add all fields instead of overwriting the map
	return func(ctx *context) {
		ctx.arguments = args
	}
}

// WithCorrelationID sets the CorrelationID
func WithCorrelationID(cID string) Option {
	return func(ctx *context) {
		ctx.correlationID = cID
	}
}

// WithParent sets the context's parent context
func WithParent(parent Context) Option {
	return func(ctx *context) {
		ctx.Context = parent
	}
}

// WithArgument sets a Context argument
func WithArgument(key string, value interface{}) Option {
	return func(ctx *context) {
		ctx.arguments[key] = value
	}
}
