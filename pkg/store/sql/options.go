package sql

type (
	Options struct {
		TTL int // minutes
	}
	Option func(*Options)
)

func WithTTL(minutes int) Option {
	return func(o *Options) {
		o.TTL = minutes
	}
}
