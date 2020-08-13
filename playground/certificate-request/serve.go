package main

import (
	"fmt"
	"os"
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/internal/daemon/config"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/net"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
)

func serve() {
	ctx := context.New()

	// load config
	fmt.Println("* loading config file")
	cfg := config.New()
	if err := cfg.Load(); err != nil {
		fmt.Println("could not load config file, error:", err)
		os.Exit(1)
	}

	// create peer key pair if it does not exist
	if cfg.Peer.PeerKey == "" {
		fmt.Println("* creating new peer key pair")
		peerKey, err := crypto.GenerateEd25519PrivateKey()
		if err != nil {
			fmt.Println("could not generate peer key, error:", err)
			os.Exit(1)
		}
		cfg.Peer.PeerKey = peerKey
	}

	// add identity key to local info
	if cfg.Peer.IdentityKey != "" {
		keychain.Put(
			keychain.IdentityKey,
			cfg.Peer.IdentityKey,
		)
	}

	// add relay peers
	for i, rp := range cfg.Peer.RelayKeys {
		eventbus.Publish(
			eventbus.RelayAdded{
				Peer: &peer.Peer{
					Owners: []crypto.PublicKey{
						crypto.PublicKey(rp),
					},
					Addresses: []string{
						cfg.Peer.RelayAddresses[i],
					},
				},
			},
		)
	}

	keychain.Put(
		keychain.PrimaryPeerKey,
		cfg.Peer.PeerKey,
	)

	// get temp bootstrap peers from cfg
	bootstrapPeers := make([]*peer.Peer, len(cfg.Peer.BootstrapKeys))
	for i, k := range cfg.Peer.BootstrapKeys {
		bootstrapPeers[i] = &peer.Peer{
			Owners: []crypto.PublicKey{
				crypto.PublicKey(k),
			},
		}
	}
	for i, a := range cfg.Peer.BootstrapAddresses {
		bootstrapPeers[i].Addresses = []string{a}
	}

	fmt.Println("* connecting to the network")

	if _, err := net.Listen(
		ctx,
		fmt.Sprintf("0.0.0.0:%d", cfg.Peer.TCPPort),
	); err != nil {
		fmt.Println("could not start listening, error:", err)
		os.Exit(1)
	}

	if err := resolver.Bootstrap(
		context.New(
			context.WithTimeout(time.Second*5),
		),
		bootstrapPeers...,
	); err != nil {
		fmt.Println("could not bootstrap, error:", err)
		os.Exit(1)
	}

	fmt.Println(
		"* ready",
		cfg.Peer.PeerKey,
		cfg.Peer.PeerKey.PublicKey(),
		net.Addresses(),
	)
}
