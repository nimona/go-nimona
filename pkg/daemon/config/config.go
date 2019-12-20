package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/caarlos0/env/v6"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
)

type APIConfig struct {
	Hostname string `json:"hostname,omitempty" env:"NIMONA_API_HOSTNAME"`
	Port     int    `json:"port,omitempty" env:"NIMONA_API_PORT"`
	Token    string `json:"token,omitempty" env:"NIMONA_API_TOKEN"`
}

type DaemonConfig struct {
	AnnounceHostname   string            `json:"hostname,omitempty" env:"NIMONA_DAEMON_HOSTNAME"`
	BootstrapAddresses []string          `json:"bootstrapAddresses,omitempty" env:"NIMONA_DAEMON_BOOTSTRAP_ADDRESSES"`
	EnableMetrics      bool              `json:"metrics,omitempty" env:"NIMONA_DAEMON_METRICS"`
	IdentityKey        crypto.PrivateKey `json:"identityKey,omitempty" env:"NIMONA_DAEMON_IDENTITY_KEY"`
	PeerKey            crypto.PrivateKey `json:"peerKey,omitempty" env:"NIMONA_DAEMON_PEER_KEY"`
	TCPPort            int               `json:"tcpPort,omitempty" env:"NIMONA_DAEMON_TCP_PORT"`
	HTTPPort           int               `json:"httpPort,omitempty" env:"NIMONA_DAEMON_HTTP_PORT"`
	RelayAddresses     []string          `json:"relayAddresses,omitempty" env:"NIMONA_DAEMON_RELAY_ADDRESSES"`
	ContentTypes       []string          `json:"contentTypes,omitempty" env:"NIMONA_CONTENT_TYPES"`
}

type Config struct {
	Path   string       `json:"-" env:"NIMONA_CONFIG" envDefault:"${HOME}/.nimona" envExpand:"true"`
	API    APIConfig    `json:"api"`
	Daemon DaemonConfig `json:"daemon"`
}

func New() *Config {
	c := &Config{
		API: APIConfig{
			Hostname: "localhost",
			Port:     10801,
		},
		Daemon: DaemonConfig{
			TCPPort:  21013,
			HTTPPort: 21083,
			BootstrapAddresses: []string{
				"tcps:egan.bootstrap.nimona.io:21013",
				"tcps:liu.bootstrap.nimona.io:21013",
				"tcps:rajaniemi.bootstrap.nimona.io:21013",
			},
		},
	}
	return c
}

func (c *Config) Load() error {
	defer func() {
		env.Parse(c) // nolint
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
