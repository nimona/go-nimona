package cmd

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"

	"nimona.io/pkg/crypto"
)

type APIConfig struct {
	Hostname string `json:"hostname,omitempty"`
	Port     int    `json:"port,omitempty"`
	Token    string `json:"token,omitempty"`
}

type DaemonConfig struct {
	AnnounceHostname   string          `json:"hostname,omitempty"`
	BootstrapAddresses []string        `json:"bootstrap_addresses,omitempty"`
	EnableMetrics      bool            `json:"metrics,omitempty"`
	IdentityKey        *crypto.Key     `json:"identity_key,omitempty"`
	Mandate            *crypto.Mandate `json:"mandate,omitempty"`
	ObjectPath         string          `json:"object_path,omitempty"`
	PeerKey            *crypto.Key     `json:"peer_key,omitempty"`
	Port               int             `json:"port,omitempty"`
	RelayAddresses     []string        `json:"relay_addresses,omitempty"`
}

type Config struct {
	API    APIConfig    `json:"api"`
	Daemon DaemonConfig `json:"daemon"`
}

func (c *Config) Update(cfgFile string) error {
	configBytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrap(err, "could not marshal config")
	}

	configFile, err := os.OpenFile(cfgFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrap(err, "could not open config")
	}

	defer configFile.Close()

	if _, err := configFile.Write(configBytes); err != nil {
		return errors.Wrap(err, "could not write config")
	}

	return nil
}
