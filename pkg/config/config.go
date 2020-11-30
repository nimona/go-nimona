package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

type Config struct {
	Path     string
	Filename string
	Peer     struct {
		PrivateKey           crypto.PrivateKey `envconfig:"PRIVATE_KEY"`
		BindAddress          string            `envconfig:"BIND_ADDRESS"`
		Bootstraps           []peer.Shorthand  `envconfig:"BOOTSTRAPS"`
		ListenOnLocalIPs     bool              `envconfig:"LISTEN_LOCAL"`
		ListenOnPrivateIPs   bool              `envconfig:"LISTEN_PRIVATE"`
		ListenOnExternalPort bool              `envconfig:"LISTEN_EXTERNAL_PORT"`
	} `envconfig:"PEER"`
	Extras map[string]interface{} `json:",omitempty"`
}

func New(opts ...Option) (*Config, error) {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}

	if err := os.MkdirAll(cfg.Path, 0700); err != nil {
		return nil, err
	}

	fullPath := filepath.Join(cfg.Path, cfg.Filename)
	configFile, err := os.OpenFile(fullPath, os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	if len(data) != 0 {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	if err := envconfig.Process("nimona", cfg); err != nil {
		return nil, err
	}

	cfg.setDefaults()

	updateData, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(fullPath, updateData, 0600); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) setDefaults() {
	if cfg.Filename == "" {
		cfg.Filename = "config.json"
	}
	if cfg.Path == "" {
		usr, _ := user.Current()
		cfg.Path = filepath.Join(usr.HomeDir, ".nimona")
	}
	if cfg.Peer.PrivateKey.IsEmpty() {
		k, _ := crypto.GenerateEd25519PrivateKey()
		cfg.Peer.PrivateKey = k
	}
	if cfg.Peer.BindAddress == "" {
		cfg.Peer.BindAddress = "0.0.0.0:0"
	}
	if len(cfg.Peer.Bootstraps) == 0 {
		cfg.Peer.Bootstraps = []peer.Shorthand{
			"ed25519.CJi6yjjXuNBFDoYYPrp697d6RmpXeW8ZUZPmEce9AgEc@tcps:asimov.bootstrap.nimona.io:22581",
			"ed25519.6fVWVAK2DVGxBhtVBvzNWNKBWk9S83aQrAqGJfrxr75o@tcps:egan.bootstrap.nimona.io:22581",
			"ed25519.7q7YpmPNQmvSCEBWW8ENw8XV8MHzETLostJTYKeaRTcL@tcps:sloan.bootstrap.nimona.io:22581",
		}
	}
}
