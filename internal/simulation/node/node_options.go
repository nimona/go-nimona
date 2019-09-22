package node

type (
	Options struct {
		Name          string
		Count         int
		Command       []string
		ContainerPort int
		NodePort      int
	}
	Option func(*Options)
)

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithCount(count int) Option {
	return func(o *Options) {
		o.Count = count
	}
}

func WithCommand(command []string) Option {
	return func(o *Options) {
		o.Command = command
	}
}

func WithNodePort(port int) Option {
	return func(o *Options) {
		o.NodePort = port
	}
}

func WithContainerPort(port int) Option {
	return func(o *Options) {
		o.ContainerPort = port
	}
}
