package config

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

type Option func(*Config)

func WithPath(path string) Option {
	return func(cfg *Config) {
		cfg.Path = path
	}
}
func WithFilename(filename string) Option {
	return func(cfg *Config) {
		cfg.Filename = filename
	}
}
func WithListenOnLocalIPs() Option {
	return func(cfg *Config) {
		cfg.Peer.ListenOnLocalIPs = true
	}
}
func WithListenOnPrivateIPs() Option {
	return func(cfg *Config) {
		cfg.Peer.ListenOnPrivateIPs = true
	}
}
func WithListenOnExternalPort() Option {
	return func(cfg *Config) {
		cfg.Peer.ListenOnExternalPort = true
	}
}

func WithDefaultPeerBindAddress(address string) Option {
	return func(cfg *Config) {
		cfg.Peer.BindAddress = address
	}
}

func WithBootstraps(peers []peer.Shorthand) Option {
	return func(cfg *Config) {
		cfg.Peer.Bootstraps = peers
	}
}

func WithPrivateKey(key crypto.PrivateKey) Option {
	return func(cfg *Config) {
		cfg.Peer.PrivateKey = key
	}
}

func WithExtraConfig(key string, data interface{}) Option {
	return func(cfg *Config) {
		//TODO: when do we do when the key already exists?
		// when to we overwrite?
		if cfg.Extras == nil {
			cfg.Extras = make(map[string]interface{})
		}
		if _, ok := cfg.Extras[key]; !ok {
			cfg.Extras[key] = data
		}
	}
}
