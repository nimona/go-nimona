package node

type (
	Options struct {
		Name         string
		Count        int
		Env          []string
		Entrypoint   []string
		Command      []string
		PortMappings map[int]int
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

func WithEnv(env []string) Option {
	return func(o *Options) {
		o.Env = env
	}
}

func WithEntrypoint(entrypoint []string) Option {
	return func(o *Options) {
		o.Entrypoint = entrypoint
	}
}

func WithCommand(command []string) Option {
	return func(o *Options) {
		o.Command = command
	}
}

func WithPortMapping(containerPort, nodePort int) Option {
	return func(o *Options) {
		o.PortMappings[containerPort] = nodePort
	}
}
