package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/caarlos0/env/v6"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/crypto"
)

type APIConfig struct {
	Hostname string `json:"hostname,omitempty" env:"NIMONA_API_HOSTNAME"`
	Port     int    `json:"port,omitempty" env:"NIMONA_API_PORT"`
	Token    string `json:"token,omitempty" env:"NIMONA_API_TOKEN"`
}

type DaemonConfig struct {
	AnnounceHostname   string             `json:"hostname,omitempty" env:"NIMONA_DAEMON_HOSTNAME"`
	BootstrapAddresses []string           `json:"bootstrap_addresses,omitempty" env:"NIMONA_DAEMON_BOOTSTRAP_ADDRESSES"`
	EnableMetrics      bool               `json:"metrics,omitempty" env:"NIMONA_DAEMON_METRICS"`
	IdentityKey        *crypto.PrivateKey `json:"identity_key,omitempty" env:"NIMONA_DAEMON_IDENTITY_KEY"`
	ObjectPath         string             `json:"object_path,omitempty" env:"NIMONA_DAEMON_OBJECT_PATH" envDefault:"${HOME}/.nimona/objects" envExpand:"true"`
	PeerKey            *crypto.PrivateKey `json:"peer_key,omitempty" env:"NIMONA_DAEMON_PEER_KEY"`
	TCPPort            int                `json:"tcp_port,omitempty" env:"NIMONA_DAEMON_TCP_PORT"`
	HTTPPort           int                `json:"http_port,omitempty" env:"NIMONA_DAEMON_HTTP_PORT"`
	RelayAddresses     []string           `json:"relay_addresses,omitempty" env:"NIMONA_DAEMON_RELAY_ADDRESSES"`
}

type Config struct {
	API    APIConfig    `json:"api"`
	Daemon DaemonConfig `json:"daemon"`
}

func LoadConfig(cfgFile string) (*Config, error) {
	c := &Config{
		API: APIConfig{
			Hostname: "localhost",
			Port:     28000,
		},
		Daemon: DaemonConfig{
			ObjectPath: path.Join(".nimona", "objects"),
			TCPPort:    21013,
			HTTPPort:   21083,
			BootstrapAddresses: []string{
				"https:andromeda.bootstrap.nimona.io:443",
				"https:borealis.bootstrap.nimona.io:443",
				"https:cassiopeia.bootstrap.nimona.io:443",
				// "https:draco.bootstrap.nimona.io:443",
				// "https:eridanus.bootstrap.nimona.io:443",
				// "https:fornax.bootstrap.nimona.io:443",
				// "https:gemini.bootstrap.nimona.io:443",
				// "https:hydra.bootstrap.nimona.io:443",
				// "https:indus.bootstrap.nimona.io:443",
				// "https:lacerta.bootstrap.nimona.io:443",
				// "https:mensa.bootstrap.nimona.io:443",
				// "https:norma.bootstrap.nimona.io:443",
				// "https:orion.bootstrap.nimona.io:443",
				// "https:pyxis.bootstrap.nimona.io:443",
				// "https:stats.bootstrap.nimona.io:443",
			},
		},
	}

	defer func() {
		env.Parse(c) // nolint
	}()

	if _, err := os.Stat(cfgFile); err != nil {
		return c, nil
	}

	jsonFile, err := os.Open(cfgFile)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close() // nolint

	jsonBytes, _ := ioutil.ReadAll(jsonFile)

	if err := json.Unmarshal(jsonBytes, c); err != nil {
		return nil, err
	}

	return c, nil
}

func UpdateConfig(cfgFile string, c *Config) error {
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
