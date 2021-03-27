package config

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"

	"github.com/mitchellh/go-homedir"
)

type Option func(*Config)

func WithoutPersistence() Option {
	return func(cfg *Config) {
		cfg.withoutPersistence = true
	}
}

func WithDefaultPath(path string) Option {
	return func(cfg *Config) {
		cfg.Path = path
		newPath, err := homedir.Expand(path)
		if err != nil {
			return
		}
		cfg.Path = newPath
	}
}

func WithDefaultFilename(filename string) Option {
	return func(cfg *Config) {
		cfg.defaultConfigFilename = filename
	}
}

func WithDefaultListenOnLocalIPs() Option {
	return func(cfg *Config) {
		cfg.Peer.ListenOnLocalIPs = true
	}
}

func WithDefaultListenOnPrivateIPs() Option {
	return func(cfg *Config) {
		cfg.Peer.ListenOnPrivateIPs = true
	}
}

func WithDefaultListenOnExternalPort() Option {
	return func(cfg *Config) {
		cfg.Peer.ListenOnExternalPort = true
	}
}

func WithDefaultDefaultPeerBindAddress(address string) Option {
	return func(cfg *Config) {
		cfg.Peer.BindAddress = address
	}
}

func WithDefaultBootstraps(peers []peer.Shorthand) Option {
	return func(cfg *Config) {
		cfg.Peer.Bootstraps = peers
	}
}

func WithDefaultPrivateKey(key crypto.PrivateKey) Option {
	return func(cfg *Config) {
		cfg.Peer.PrivateKey = key
	}
}

func WithExtraConfig(key string, data interface{}) Option {
	return func(cfg *Config) {
		if cfg.extras == nil {
			cfg.extras = make(map[string]interface{})
		}
		if _, ok := cfg.extras[key]; !ok {
			cfg.extras[key] = data
		}
	}
}
