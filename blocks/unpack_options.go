package blocks

func ParseUnpackOptions(opts ...UnpackOption) *UnpackOptions {
	options := &UnpackOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

type UnpackOptions struct {
	Verify             bool
	DecodeNested       bool
	DecodeNestedBase58 bool
}

type UnpackOption func(*UnpackOptions)

func DecodeNested() UnpackOption {
	return func(opts *UnpackOptions) {
		opts.DecodeNested = true
	}
}

func Verify() UnpackOption {
	return func(opts *UnpackOptions) {
		opts.Verify = true
	}
}

func DecodeNestedBase58() UnpackOption {
	return func(opts *UnpackOptions) {
		opts.DecodeNestedBase58 = true
	}
}
