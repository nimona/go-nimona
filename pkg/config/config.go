package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/stoewer/go-strcase"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

type (
	Config struct {
		Path     string `json:"-"`
		LogLevel string `json:"logLevel" envconfig:"LOG_LEVEL"`
		Peer     struct {
			PrivateKey           crypto.PrivateKey `json:"privateKey" envconfig:"PRIVATE_KEY"`
			BindAddress          string            `json:"bindAddress" envconfig:"BIND_ADDRESS"`
			Bootstraps           []peer.Shorthand  `json:"bootstraps" envconfig:"BOOTSTRAPS"`
			ListenOnLocalIPs     bool              `json:"listenLocalIPs" envconfig:"LISTEN_LOCAL"`
			ListenOnPrivateIPs   bool              `json:"listenPrivateIPs" envconfig:"LISTEN_PRIVATE"`
			ListenOnExternalPort bool              `json:"listenExternalPort" envconfig:"LISTEN_EXTERNAL_PORT"`
		} `json:"peer" envconfig:"PEER"`
		Extras map[string]json.RawMessage `json:"extras,omitempty"`
		extras map[string]interface{}
		// internal defaults
		defaultConfigFilename string
	}
)

func New(opts ...Option) (*Config, error) {
	currentUser, _ := user.Current()

	cfg := &Config{
		Path:                  filepath.Join(currentUser.HomeDir, ".nimona"),
		Extras:                map[string]json.RawMessage{},
		defaultConfigFilename: "config.json",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if configDir := os.Getenv("NIMONA_CONFIG_DIR"); configDir != "" {
		cfg.Path = configDir
	}

	if configFilename := os.Getenv("NIMONA_CONFIG_FILE"); configFilename != "" {
		cfg.defaultConfigFilename = configFilename
	}

	if err := os.MkdirAll(cfg.Path, 0700); err != nil {
		return nil, fmt.Errorf("error creating directory, %w", err)
	}

	fullPath := filepath.Join(cfg.Path, cfg.defaultConfigFilename)

	configFile, err := os.OpenFile(fullPath, os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("error opening file, %w", err)
	}

	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file, %w", err)
	}

	if len(data) != 0 {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("error unmarshaling file, %w", err)
		}
		for k, r := range cfg.Extras {
			target, ok := cfg.extras[k]
			if !ok {
				continue
			}
			if err := json.Unmarshal(r, target); err != nil {
				return nil, fmt.Errorf("error unmarshaling extras, %w", err)
			}
		}
	}

	cfg.setDefaults()

	for k, r := range cfg.extras {
		data, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshaling extras, %w", err)
		}
		cfg.Extras[strcase.LowerCamelCase(k)] = data
	}

	updateData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling config, %w", err)
	}

	if err := envconfig.Process("nimona", cfg); err != nil {
		return nil, fmt.Errorf("error processing env vars, %w", err)
	}

	for k, r := range cfg.extras {
		if err := envconfig.Process("nimona_"+k, r); err != nil {
			return nil, fmt.Errorf("error processing extra env vars, %w", err)
		}
	}

	if err := ioutil.WriteFile(fullPath, updateData, 0600); err != nil {
		return nil, fmt.Errorf("error writing file, %w", err)
	}

	return cfg, nil
}

func (cfg *Config) setDefaults() {
	if cfg.Peer.PrivateKey.IsEmpty() {
		k, _ := crypto.NewEd25519PrivateKey(crypto.PeerKey)
		cfg.Peer.PrivateKey = k
	}
	if cfg.Peer.BindAddress == "" {
		cfg.Peer.BindAddress = "0.0.0.0:0"
	}
	if len(cfg.Peer.Bootstraps) == 0 {
		cfg.Peer.Bootstraps = []peer.Shorthand{
			"bahwqdag4aeqewwlutsgr7kv2iaqsrnppbdcmyykpckqn5uaqczae6fergklclea@tcps:asimov.bootstrap.nimona.io:22581", // nolint: lll
			"bahwqdag4aeqomor45il7jjxlox7y5aj6cigawcljgsfftytwf6ulrpfqtiuzsya@tcps:egan.bootstrap.nimona.io:22581",   // nolint: lll
			"bahwqdag4aeqm5gkdk7dlbzke6wgc7rkm67cnqiv2jctfoxoo3vjmbdpjt5qi6za@tcps:sloan.bootstrap.nimona.io:22581",  // nolint: lll
		}
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "FATAL"
	}
}
