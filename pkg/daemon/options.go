package daemon

import "nimona.io/pkg/config"

func WithConfigOptions(opts ...config.Option) Option {
	return func(d *daemon) error {
		d.configOptions = opts
		return nil
	}
}
