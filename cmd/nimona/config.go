package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/caarlos0/env/v6"

	"nimona.io/internal/errors"
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
				// "tcps:andromeda.nimona.io:21013",
				// "tcps:borealis.nimona.io:21013",
				"tcps:cassiopeia.nimona.io:21013",
				// "tcps:draco.nimona.io:21013",
				// "tcps:eridanus.nimona.io:21013",
				// "tcps:fornax.nimona.io:21013",
				// "tcps:gemini.nimona.io:21013",
				// "tcps:hydra.nimona.io:21013",
				// "tcps:indus.nimona.io:21013",
				// "tcps:lacerta.nimona.io:21013",
				// "tcps:mensa.nimona.io:21013",
				// "tcps:norma.nimona.io:21013",
				// "tcps:orion.nimona.io:21013",
				// "tcps:pyxis.nimona.io:21013",
				// "tcps:stats.nimona.io:21013",
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

	defer configFile.Close()

	if _, err := configFile.Write(configBytes); err != nil {
		return errors.Wrap(err, errors.New("could not write config"))
	}

	return nil
}
