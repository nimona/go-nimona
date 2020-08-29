package main

import (
	"fmt"
	"os"
	"time"

	"nimona.io/internal/daemon/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
)

func serve() network.Network {
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

	local := localpeer.New()

	// add identity key to local info
	if cfg.Peer.IdentityKey != "" {
		local.PutPrimaryIdentityKey(cfg.Peer.IdentityKey)
	}

	// add relay peers
	for i, rp := range cfg.Peer.RelayKeys {
		local.PutRelays(
			&peer.Peer{
				Metadata: object.Metadata{
					Owner: crypto.PublicKey(rp),
				},
				Addresses: []string{
					cfg.Peer.RelayAddresses[i],
				},
			},
		)
	}

	local.PutPrimaryPeerKey(cfg.Peer.PeerKey)

	// get temp bootstrap peers from cfg
	bootstrapPeers := make([]*peer.Peer, len(cfg.Peer.BootstrapKeys))
	for i, k := range cfg.Peer.BootstrapKeys {
		bootstrapPeers[i] = &peer.Peer{
			Metadata: object.Metadata{
				Owner: crypto.PublicKey(k),
			},
		}
	}
	for i, a := range cfg.Peer.BootstrapAddresses {
		bootstrapPeers[i].Addresses = []string{a}
	}

	fmt.Println("* connecting to the network")

	net := network.New(ctx, network.WithLocalPeer(local))
	if _, err := net.Listen(
		ctx,
		fmt.Sprintf("0.0.0.0:%d", cfg.Peer.TCPPort),
	); err != nil {
		fmt.Println("could not start listening, error:", err)
		os.Exit(1)
	}

	res := resolver.New(
		ctx,
		net,
	)

	if err := res.Bootstrap(
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
		net.LocalPeer().GetAddresses(),
	)

	return net
}
