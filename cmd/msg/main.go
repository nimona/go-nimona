package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"nimona.io/internal/daemon/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/network"
	"nimona.io/pkg/localpeer"
	"nimona.io/internal/nat"
	"nimona.io/internal/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
)

func main() {
	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)

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

	// create identity key pair if it does not exist
	// TODO this is temporary
	if cfg.Peer.IdentityKey == "" {
		fmt.Println("* creating new identity key pair")
		identityKey, err := crypto.GenerateEd25519PrivateKey()
		if err != nil {
			fmt.Println("could not generate identity key, error:", err)
			os.Exit(1)
		}
		cfg.Peer.IdentityKey = identityKey
	}

	// update config
	if err := cfg.Update(); err != nil {
		fmt.Println("could not update config, error:", err)
		os.Exit(1)
	}

	// add identity key to local info
	if cfg.Peer.IdentityKey != "" {
		localpeer.Put(
			localpeer.PrimaryIdentityKey,
			cfg.Peer.IdentityKey,
		)
	}

	// add relay peers
	for i, rp := range cfg.Peer.RelayKeys {
		eventbus.Publish(
			eventbus.RelayAdded{
				Peer: &peer.Peer{
					Metadata: object.Metadata{
						Owner: crypto.PublicKey(rp),
					},
					Addresses: []string{
						cfg.Peer.RelayAddresses[i],
					},
				},
			},
		)
	}

	localpeer.Put(
		localpeer.PrimaryPeerKey,
		cfg.Peer.PeerKey,
	)

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

	if _, err := net.Listen(
		ctx,
		fmt.Sprintf("0.0.0.0:%d", cfg.Peer.TCPPort),
	); err != nil {
		fmt.Println("could not start listening, error:", err)
		os.Exit(1)
	}

	res := resolver.New(
		ctx,
		resolver.WithEventbus(eventbus.DefaultEventbus),
		resolver.WithExchange(exchange.DefaultExchange),
		resolver.WithLocalPeer(localpeer.DefaultLocalPeer),
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

	rmBinding, err := nat.MapExternalPort(cfg.Peer.TCPPort)
	if err != nil {
		fmt.Println("* could not create UPNP mapping, error:", err)
	}
	defer rmBinding()

	sub := exchange.Subscribe(
		exchange.FilterByObjectType("nimona.io/msg"),
	)
	go func() {
		for {
			e, err := sub.Next()
			if err != nil {
				panic(err)
			}
			fmt.Println(">", e.Sender.String(), e.Payload.Get("body:s").(string))
		}
	}()

	fmt.Println("* ready")
	fmt.Println("* type `info` to see your peer's information.")
	fmt.Println("* type `send <recipient's key> <message>` to send a message.")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		cmd := strings.Split(text, " ")[0]
		switch cmd {
		case "":
			continue
		case "info":
			fmt.Println("> addresses", net.Addresses())
			fmt.Println("> peer", cfg.Peer.PeerKey.PublicKey().String())
			fmt.Println("> identity", cfg.Peer.IdentityKey.PublicKey().String())
		case "exit", "quit":
			fmt.Println("> goodbye 👋")
			return
		case "send":
			ps := strings.Split(text, " ")
			if len(ps) < 3 {
				fmt.Println("missing recipient")
				continue
			}
			recipient := ps[1]
			fmt.Println("* looking for ", recipient)
			body := strings.Join(ps[2:], " ")
			msg := Msg{
				Datetime: time.Now().Unix(),
				Body:     body,
			}
			rs, err := res.Lookup(
				context.New(
					context.WithTimeout(time.Second),
				),
				resolver.LookupByOwner(crypto.PublicKey(recipient)),
			)
			if err != nil {
				fmt.Println("* could not find recipient, error:, error:", err)
				break
			}
			sent := 0
			failed := 0
			for r := range rs {
				fmt.Println("** found on", r.Addresses)
				if err := exchange.Send(
					context.New(
						context.WithTimeout(time.Second),
					),
					msg.ToObject(),
					r,
				); err != nil {
					failed++
					fmt.Println("* could not send message;, error:", err)
					continue
				}
				sent++
			}
			if sent > 0 {
				fmt.Println("* message sent")
			} else if sent == 0 && failed == 0 {
				fmt.Println("* could not find recipient")
			} else {
				fmt.Println("* could not send message")
			}
		default:
			fmt.Println("* not sure how to handle", cmd)
		}
	}
}
