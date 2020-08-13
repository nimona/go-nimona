package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/kelseyhightower/envconfig"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
)

type APIConfig struct {
	Host  string `json:"host,omitempty" envconfig:"HOST"`
	Port  int    `json:"port,omitempty" envconfig:"PORT"`
	Token string `json:"token,omitempty" envconfig:"TOKEN"`
}

// nolint: lll
type PeerConfig struct {
	AnnounceHostname   string            `json:"hostname,omitempty" envconfig:"HOSTNAME"`
	BootstrapKeys      []string          `json:"bootstrapKeys,omitempty" envconfig:"BOOTSTRAP_KEYS"`
	BootstrapAddresses []string          `json:"bootstrapAddresses,omitempty" envconfig:"BOOTSTRAP_ADDRESSES"`
	EnableMetrics      bool              `json:"metrics,omitempty" envconfig:"METRICS"`
	EnableDebug        bool              `json:"debug,omitempty" envconfig:"DEBUG"`
	IdentityKey        crypto.PrivateKey `json:"identityKey,omitempty" envconfig:"IDENTITY_KEY"`
	PeerKey            crypto.PrivateKey `json:"peerKey,omitempty" envconfig:"KEY"`
	TCPPort            int               `json:"tcpPort,omitempty" envconfig:"TCP_PORT"`
	UPNP               bool              `json:"upnp,omitempty" envconfig:"UPNP"`
	RelayKeys          []string          `json:"relayKeys,omitempty" envconfig:"RELAY_KEYS"`
	RelayAddresses     []string          `json:"relayAddresses,omitempty" envconfig:"RELAY_ADDRESSES"`
	ContentTypes       []string          `json:"contentTypes,omitempty" envconfig:"CONTENT_TYPES"`
}

type Config struct {
	Path string     `json:"-"`
	API  APIConfig  `json:"api"`
	Peer PeerConfig `json:"peer"`
}

func New() *Config {
	c := &Config{
		API: APIConfig{
			Host: "localhost",
			Port: 10801,
		},
		Peer: PeerConfig{
			TCPPort: 21013,
			BootstrapKeys: []string{
				"ed25519.Eq4y6LPB1cHX5pM8rgK9vbjFSt6e6hDW8rAwTu1TUaXW",
				"ed25519.7GhQzPptXaBehoaq3DfcSfj1Z7X5tyjinEdx1R7mqfx8",
				"ed25519.B6KNw8oyerJRpPKtHrf5YSCeBqELbAfWScxRLNdGbazG",
			},
			BootstrapAddresses: []string{
				"tcps:egan.bootstrap.nimona.io:21013",
				"tcps:liu.bootstrap.nimona.io:21013",
				"tcps:rajaniemi.bootstrap.nimona.io:21013",
			},
			RelayKeys: []string{
				"ed25519.Eq4y6LPB1cHX5pM8rgK9vbjFSt6e6hDW8rAwTu1TUaXW",
				"ed25519.7GhQzPptXaBehoaq3DfcSfj1Z7X5tyjinEdx1R7mqfx8",
				"ed25519.B6KNw8oyerJRpPKtHrf5YSCeBqELbAfWScxRLNdGbazG",
			},
			RelayAddresses: []string{
				"tcps:egan.bootstrap.nimona.io:21013",
				"tcps:liu.bootstrap.nimona.io:21013",
				"tcps:rajaniemi.bootstrap.nimona.io:21013",
			},
		},
	}
	return c
}

func (c *Config) Load() error {
	if p := os.Getenv("NIMONA_CONFIG"); p != "" {
		c.Path = p
	}
	c.Path = os.ExpandEnv(c.Path)

	defer func() {
		envconfig.Process("nimona_api", &c.API)   // nolint: errcheck
		envconfig.Process("nimona_peer", &c.Peer) // nolint: errcheck
	}()

	cfgFile := path.Join(c.Path, "config.json")

	if _, err := os.Stat(cfgFile); err != nil {
		return nil
	}

	jsonFile, err := os.Open(cfgFile)
	if err != nil {
		return err
	}

	defer jsonFile.Close() // nolint

	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(jsonBytes, c); err != nil {
		return err
	}

	return nil
}

func (c *Config) Update() error {
	cfgFile := path.Join(c.Path, "config.json")
	if _, err := os.Stat(cfgFile); err != nil {
		basepath := path.Dir(cfgFile)
		if err := os.MkdirAll(basepath, 0777); err != nil {
			return err
		}
	}

	configBytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrap(err, errors.New("could not marshal config"))
	}

	configFile, err := os.OpenFile(cfgFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrap(err, errors.New("could not open config"))
	}

	defer configFile.Close() // nolint: errcheck

	configFile.Truncate(0) // nolint: errcheck

	if _, err := configFile.Write(configBytes); err != nil {
		return errors.Wrap(err, errors.New("could not write config"))
	}

	return nil
}