package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kelseyhightower/envconfig"

	"nimona.io"
)

type Config struct {
	AddressPublic  string `envconfig:"ADDRESS_PUBLIC"`
	StoreType      string `envconfig:"STORE_TYPE"`
	StoreHostname  string `envconfig:"STORE_HOSTNAME"`
	StorePort      string `envconfig:"STORE_PORT"`
	StoreUsername  string `envconfig:"STORE_USERNAME"`
	StorePassword  string `envconfig:"STORE_PASSWORD"`
	PeerPublicKey  string `envconfig:"PEER_PUBLIC_KEY"`
	PeerPrivateKey string `envconfig:"PEER_PRIVATE_KEY"`
}

func main() {
	cfg := &Config{}
	err := envconfig.Process("nimona", cfg)
	if err != nil {
		panic(fmt.Errorf("error parsing config: %w", err))
	}

	var publicKey nimona.PublicKey
	var privateKey nimona.PrivateKey

	if cfg.PeerPublicKey == "" || cfg.PeerPrivateKey == "" {
		publicKey, privateKey, err = nimona.GenerateKey()
		if err != nil {
			panic(fmt.Errorf("error generating key: %w", err))
		}
	} else {
		publicKey, err = nimona.ParsePublicKey(cfg.PeerPublicKey)
		if err != nil {
			panic(fmt.Errorf("error parsing public key: %w", err))
		}
		privateKey, err = nimona.ParsePrivateKey(cfg.PeerPrivateKey)
		if err != nil {
			panic(fmt.Errorf("error parsing private key: %w", err))
		}
	}

	ctx := context.Background()
	transport := &nimona.TransportUTP{}
	listener, err := transport.Listen(ctx, "0.0.0.0:1013")
	if err != nil {
		panic(fmt.Errorf("error listening: %w", err))
	}

	// docDB, err := gorm.Open(
	// 	sqlite.Open("file::memory:?cache=shared"),
	// 	&gorm.Config{},
	// )
	// if err != nil {
	// 	panic(fmt.Errorf("error opening db: %w", err))
	// }

	// docStore, err := nimona.NewDocumentStore(docDB)
	// if err != nil {
	// 	panic(fmt.Errorf("error creating document store: %w", err))
	// }

	sesManager, err := nimona.NewSessionManager(
		transport,
		listener,
		publicKey,
		privateKey,
	)
	if err != nil {
		panic(fmt.Errorf("error creating session manager: %w", err))
	}

	peerInfo := &nimona.PeerInfo{
		PublicKey: publicKey,
		Addresses: []nimona.PeerAddr{{
			PublicKey: publicKey,
			Address:   listener.PeerAddr().Address,
			Network:   listener.PeerAddr().Network,
		}},
	}

	peerConfig := nimona.NewPeerConfig(privateKey, publicKey, peerInfo)

	pingHandler := &nimona.HandlerPing{
		PeerConfig: peerConfig,
	}

	sesManager.RegisterHandler(
		"test/ping",
		pingHandler.HandlePingRequest,
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, peerConfig.GetPeerInfo().Addresses[0].String())
	})

	http.ListenAndServe("0.0.0.0:80", nil) // nolint: errcheck
}
