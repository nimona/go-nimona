package config

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

type Option func(*Config, *otherOtions)

func WithDefaultPath(path string) Option {
	return func(cfg *Config, oopts *otherOtions) {
		cfg.Path = path
	}
}

func WithDefaultListenOnLocalIPs() Option {
	return func(cfg *Config, oopts *otherOtions) {
		cfg.Peer.ListenOnLocalIPs = true
	}
}

func WithDefaultListenOnPrivateIPs() Option {
	return func(cfg *Config, oopts *otherOtions) {
		cfg.Peer.ListenOnPrivateIPs = true
	}
}

func WithDefaultListenOnExternalPort() Option {
	return func(cfg *Config, oopts *otherOtions) {
		cfg.Peer.ListenOnExternalPort = true
	}
}

func WithDefaultDefaultPeerBindAddress(address string) Option {
	return func(cfg *Config, oopts *otherOtions) {
		cfg.Peer.BindAddress = address
	}
}

func WithDefaultBootstraps(peers []peer.Shorthand) Option {
	return func(cfg *Config, oopts *otherOtions) {
		cfg.Peer.Bootstraps = peers
	}
}

func WithDefaultPrivateKey(key crypto.PrivateKey) Option {
	return func(cfg *Config, oopts *otherOtions) {
		cfg.Peer.PrivateKey = key
	}
}

func WithExtraConfig(key string, data interface{}) Option {
	return func(cfg *Config, oopts *otherOtions) {
		if cfg.Extras == nil {
			cfg.Extras = make(map[string]interface{})
		}
		if _, ok := cfg.Extras[key]; !ok {
			cfg.Extras[key] = data
		}
	}
}

func WithAdditionalEnvVars(pairs map[string]string) Option {
	return func(cfg *Config, oopts *otherOtions) {
		oopts.additionalPairs = pairs
	}
}
