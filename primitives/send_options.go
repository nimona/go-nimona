package primitives

func ParseSendOptions(opts ...SendOption) *SendOptions {
	options := &SendOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

type SendOptions struct {
	Sign            bool
	SignWithMandate bool
}

type SendOption func(*SendOptions)

func SendOptionSignWithMandate() SendOption {
	return func(opts *SendOptions) {
		opts.SignWithMandate = true
	}
}

func SendOptionSign() SendOption {
	return func(opts *SendOptions) {
		opts.Sign = true
	}
}
