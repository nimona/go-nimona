package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/nimona/go-nimona/api"
	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/codec"
	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/net"
	"github.com/nimona/go-nimona/peers"
	"github.com/nimona/go-nimona/storage"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	bootstrapPeerInfos = []string{
		"WUvH4xVzmVzstkzrrE6jv5F6usGHWGwfWR3euPvM4w5WRzjpL2LjxVSzJEvkVzW7zJE2aDgUtoKwAYP6rBmJSpiYEs1zZFL6rR6LQUpf9M45EHTZqK2AFQSMgDxyRrFweZ7QDqU64uSGr1s7ZRKhbzPHBz8HrTSczH7E5kHaLF3aqAHkK4PaHM3yAaTnzj2Tz7QC4GADzBAopFXGzptdZmwdBE9inZX8amJNptkvjGtSKW4FyaafuTbUCsuFaWweiXfNy1hVVdiKBdc4uz19B6ZN4HSnvQE1a9yw62v9gPr2ULFwU5pXPd3KDr3UFBEAsj4idJQjVxg7wckrdsceB5TqMYRx8JSQHr3spbVXC1ZZUx2SsBWLqzLFjzHsb374upxBkaiA",
	}
)

func base64ToBytes(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// Hello payload
type Hello struct {
	Body string
}

func init() {
	// telemetry.SetupKeenCollector()
	blocks.RegisterContentType("demo.hello", Hello{}, blocks.Persist())
}

func decodeSignature(sig string) *blocks.Signature {
	bytes, err := blocks.Base58Decode(sig)
	if err != nil {
		panic(err)
	}

	signature := &blocks.Signature{}
	if err := codec.Unmarshal(bytes, signature); err != nil {
		panic(err)
	}

	return signature
}

func main() {
	configPath := os.Getenv("NIMONA_PATH")

	if configPath == "" {
		usr, _ := user.Current()
		configPath = path.Join(usr.HomeDir, ".nimona")
	}

	if err := os.MkdirAll(configPath, 0777); err != nil {
		log.Fatal("could not create config dir", err)
	}

	reg, err := peers.NewAddressBook(configPath)
	if err != nil {
		log.Fatal("could not load key", err)
	}

	port, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 32)

	for _, peerInfoB58 := range bootstrapPeerInfos {
		peerInfoBytes, _ := blocks.Base58Decode(peerInfoB58)
		peerInfo, err := blocks.Unmarshal(peerInfoBytes)
		if err != nil {
			panic(err)
		}
		if err := reg.PutPeerInfo(peerInfo.(*peers.PeerInfo)); err != nil {
			log.Fatal("could not put bootstrap peer", err)
		}
	}

	storagePath := path.Join(configPath, "storage")

	dpr := storage.NewDiskStorage(storagePath)
	n, _ := net.NewExchange(reg, dpr)
	dht, _ := dht.NewDHT(n, reg)
	n.RegisterDiscoverer(dht)

	n.Listen(context.Background(), fmt.Sprintf("0.0.0.0:%d", port))

	n.Handle("demo", func(payload interface{}) error {
		fmt.Printf("___ Got block %s\n", payload.(*Hello).Body)
		return nil
	})

	httpPort := "26880"
	if nhp := os.Getenv("HTTP_PORT"); nhp != "" {
		httpPort = nhp
	}
	httpAddress := ":" + httpPort
	api := api.New(reg, dht, dpr)
	go api.Serve(httpAddress)

	shell := ishell.New()
	shell.Printf("Nimona DHT (%s)\n", version)

	putProvider := &ishell.Cmd{
		Name:    "providers",
		Aliases: []string{"provider"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) < 1 {
				c.Println("Missing providing key")
				return
			}
			key := c.Args[0]
			ctx := context.Background()
			if err := dht.PutProviders(ctx, key); err != nil {
				c.Printf("Could not put key %s\n", key)
				c.Printf("Error: %s\n", err)
			}
		},
		Help: "announce a provided key on the dht",
	}

	getProvider := &ishell.Cmd{
		Name:    "providers",
		Aliases: []string{"provider"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) == 0 {
				c.Println("Missing key")
				return
			}
			c.ProgressBar().Indeterminate(true)
			c.ProgressBar().Start()
			key := c.Args[0]
			ctx := context.Background()
			rs, err := dht.GetProviders(ctx, key)
			c.Println("")
			if err != nil {
				c.Printf("Could not get providers for key %s\n", key)
				c.Printf("Error: %s\n", err)
			}
			providers := []string{}
			for provider := range rs {
				providers = append(providers, provider.Thumbprint())
			}
			c.Println("* " + key)
			for _, peerID := range providers {
				c.Printf("  - %s\n", peerID)
			}
			c.ProgressBar().Stop()
		},
		Help: "get peers providing a value from the dht",
	}

	getBlock := &ishell.Cmd{
		Name:    "blocks",
		Aliases: []string{"block"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) < 1 {
				c.Println("Missing block id")
				return
			}

			ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
			defer cf()

			block, err := n.Get(ctx, c.Args[0])
			if err != nil {
				c.Println(err)
				return
			}

			c.Printf("Received block %#v\n", block)
		},
		Help: "get peers providing a value from the dht",
	}

	listProviders := &ishell.Cmd{
		Name:    "providers",
		Aliases: []string{"provider"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := dht.GetAllProviders()
			for key, vals := range ps {
				c.Println("* " + key)
				for _, val := range vals {
					c.Printf("  - %s\n", val)
				}
			}
		},
		Help: "list all providers stored in our local dht",
	}

	listPeers := &ishell.Cmd{
		Name:    "peers",
		Aliases: []string{"peer"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := reg.GetAllPeerInfo()
			for _, peer := range ps {
				c.Println("* " + peer.Thumbprint())
				c.Printf("  - addresses:\n")
				for _, address := range peer.Addresses {
					c.Printf("     - %s\n", address)
				}
			}
		},
		Help: "list all peers stored in our local dht",
	}

	listBlocks := &ishell.Cmd{
		Name: "blocks",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			blocks, err := n.GetLocalBlocks()
			if err != nil {
				c.Println(err)
				return
			}
			for _, block := range blocks {
				c.Printf("     - %s\n", block)
			}
		},
		Help: "list all blocks in local storage",
	}

	listLocal := &ishell.Cmd{
		Name: "local",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			peer := reg.GetLocalPeerInfo()
			c.Println("* " + peer.Thumbprint())
			c.Printf("  - addresses:\n")
			for _, address := range peer.Addresses {
				c.Printf("     - %s\n", address)
			}
		},
		Help: "list protocols for local peer",
	}

	send := &ishell.Cmd{
		Name: "send",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Println("Missing peer id or block")
				return
			}
			ctx := context.Background()
			msg := strings.Join(c.Args[1:], " ")
			peer, err := reg.GetPeerInfo(c.Args[0])
			if err != nil {
				c.Println("Could not get peer")
				return
			}
			signer := reg.GetLocalPeerInfo().Key
			if err := n.Send(ctx, &Hello{Body: msg}, peer.Signature.Key, blocks.SignWith(signer)); err != nil {
				c.Println("Could not send block", err)
				return
			}
		},
		Help: "list protocols for local peer",
	}

	block := &ishell.Cmd{
		Name: "block",
		Help: "send blocks to peers",
	}

	get := &ishell.Cmd{
		Name: "get",
		Help: "get resource",
	}

	get.AddCmd(getProvider)
	get.AddCmd(getBlock)

	put := &ishell.Cmd{
		Name: "put",
		Help: "put resource",
	}

	put.AddCmd(putProvider)

	list := &ishell.Cmd{
		Name:    "list",
		Aliases: []string{"l", "ls"},
		Help:    "list cached resources",
	}

	list.AddCmd(listProviders)
	list.AddCmd(listPeers)
	list.AddCmd(listLocal)
	list.AddCmd(listBlocks)

	shell.AddCmd(block)
	shell.AddCmd(get)
	shell.AddCmd(put)
	shell.AddCmd(list)
	shell.AddCmd(send)

	// when started with "exit" as first argument, assume non-interactive execution
	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		// start shell
		shell.Run()
		// teardown
		shell.Close()
	}
}
