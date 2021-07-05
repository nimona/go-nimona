package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/go-homedir"
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
		withoutPersistence    bool
	}
)

func New(opts ...Option) (*Config, error) {
	cfg := &Config{
		Path:                  ".nimona",
		Extras:                map[string]json.RawMessage{},
		defaultConfigFilename: "config.json",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if configDir := os.Getenv("NIMONA_CONFIG_DIR"); configDir != "" {
		cfg.Path = configDir
	}

	newPath, err := homedir.Expand(cfg.Path)
	if err != nil {
		return nil, err
	}
	cfg.Path = newPath

	if configFilename := os.Getenv("NIMONA_CONFIG_FILE"); configFilename != "" {
		cfg.defaultConfigFilename = configFilename
	}

	if err := os.MkdirAll(cfg.Path, 0700); err != nil {
		return nil, fmt.Errorf("error creating directory, %w", err)
	}

	if cfg.withoutPersistence {
		cfg.setDefaults()
		return cfg, nil
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
		k, _ := crypto.NewEd25519PrivateKey()
		cfg.Peer.PrivateKey = k
	}
	if cfg.Peer.BindAddress == "" {
		cfg.Peer.BindAddress = "0.0.0.0:0"
	}
	if len(cfg.Peer.Bootstraps) == 0 {
		cfg.Peer.Bootstraps = []peer.Shorthand{
			"z6MkvTRseMpsTW3Knm5LJYmQU7JVjZx3gEceNGRgT8cNsf5t@tcps:asimov.bootstrap.nimona.io:22581", // nolint: lll
			"z6MkvY55pieg8jUfyhYtv6YEmLqCsEgF8TY4QeCufHJYBnmi@tcps:egan.bootstrap.nimona.io:22581",   // nolint: lll
			"z6MkjjHRY3jJKiWNFLULdYQUWfP8ASZmUrEnNmBtUhAGNxGB@tcps:sloan.bootstrap.nimona.io:22581",  // nolint: lll
		}
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "FATAL"
	}
}
