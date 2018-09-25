package primitives

func ParseSendOptions(opts ...SendOption) *SendOptions {
	options := &SendOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

type SendOptions struct {
	Key  *Key
	Sign bool
}

type SendOption func(*SendOptions)

func SignWith(key *Key) SendOption {
	return func(opts *SendOptions) {
		opts.Key = key
		opts.Sign = true
	}
}
