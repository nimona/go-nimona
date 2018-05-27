package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/nimona/go-nimona/api"
	"github.com/nimona/go-nimona/blx"
	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/wire"

	"gopkg.in/abiosoft/ishell.v2"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var bootstrapPeerInfos = []mesh.PeerInfo{
	mesh.PeerInfo{
		ID: "7730b73e34ae2e3ad92235aefc7ee0366736602f96785e6f35e8b710923b4562",
		Addresses: []string{
			"tcp:andromeda.nimona.io:26801",
		},
		PublicKey: [32]byte{
			119, 48, 183, 62, 52, 174, 46, 58, 217, 34, 53, 174, 252, 126,
			224, 54, 103, 54, 96, 47, 150, 120, 94, 111, 53, 232, 183, 16,
			146, 59, 69, 98,
		},
	},
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

	keyPath := path.Join(configPath, "peer.json")
	spi, err := mesh.LoadOrCreateLocalPeerInfo(keyPath)
	if err != nil {
		log.Fatal("could not load key", err)
	}

	port, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 32)

	reg := mesh.NewRegisty(spi)

	for _, peerInfo := range bootstrapPeerInfos {
		reg.PutPeerInfo(&peerInfo)
	}

	storagePath := path.Join(configPath, "storage")

	wre, _ := wire.NewWire(reg)
	dht, _ := dht.NewDHT(wre, reg)
	dpr := blx.NewDiskStorage(storagePath)
	blx, _ := blx.NewBlockExchange(wre, dpr)

	// Announce blocks on init and on new blocks
	go func() {
		blockKeys, _ := blx.GetLocalBlocks()

		for _, bk := range blockKeys {
			// TODO Check errors
			dht.PutProviders(context.Background(), bk)
		}

		// TODO Store the unsubscribe key
		blx.Subscribe(func(key string) {
			dht.PutProviders(context.Background(), key)
		})
	}()

	wre.Listen(fmt.Sprintf("0.0.0.0:%d", port))

	wre.HandleExtensionEvents("msg", func(event *wire.Message) error {
		fmt.Printf("___ Got message %s\n", string(event.Payload))
		return nil
	})

	httpPort := "26880"
	if nhp := os.Getenv("HTTP_PORT"); nhp != "" {
		httpPort = nhp
	}
	httpAddress := ":" + httpPort
	api := api.New(reg, dht)
	go api.Serve(httpAddress)

	shell := ishell.New()
	shell.Printf("Nimona DHT (%s)\n", version)

	putValue := &ishell.Cmd{
		Name:    "values",
		Aliases: []string{"value"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) < 2 {
				c.Println("Missing key and value")
				return
			}

			key := c.Args[0]
			val := strings.Join(c.Args[1:], " ")
			ctx := context.Background()
			if err := dht.PutValue(ctx, key, val); err != nil {
				c.Printf("Could not put key %s\n", key)
				c.Printf("Error: %s\n", err)
			}
		},
		Help: "put a value on the dht",
	}

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

	getValue := &ishell.Cmd{
		Name:    "values",
		Aliases: []string{"value"},
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
			rs, err := dht.GetValue(ctx, key)
			c.Println("")
			if err != nil {
				c.Printf("Could not get %s\n", key)
				c.Printf("Error: %s\n", err)
			}
			if rs != "" {
				c.Printf(" - %s\n", rs)
			}
			c.ProgressBar().Stop()
		},
		Help: "get a value from the dht",
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
			c.Println("* " + key)
			for _, peerID := range rs {
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
				c.Println("Missing key peer")
				return
			}

			peer := ""

			if len(c.Args) == 2 {
				peer = c.Args[1]
			}

			blockHash := c.Args[0]

			block, err := blx.Get(blockHash, peer)
			if err != nil {
				c.Println(err)
				return
			}

			c.Printf("Received block of %d bytes length\n", len(block.Data))
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

	listValues := &ishell.Cmd{
		Name:    "values",
		Aliases: []string{"value"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := dht.GetAllValues()
			for key, val := range ps {
				c.Printf("* %s: %s\n", key, val)
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
				c.Println("* " + peer.ID)
				c.Printf("  - public key: %x\n", peer.PublicKey)
				c.Printf("  - signature: %x\n", peer.Signature)
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

			blocks, err := blx.GetLocalBlocks()
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
			c.Println("* " + peer.ID)
			c.Printf("  - public key: %x\n", peer.PublicKey)
			c.Printf("  - signature: %x\n", peer.Signature)
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
				c.Println("Missing peer id or message")
				return
			}
			ctx := context.Background()
			msg := strings.Join(c.Args[1:], " ")
			to := []string{c.Args[0]}
			if err := wre.Send(ctx, "msg", "msg", msg, to); err != nil {
				c.Println("Could not send message", err)
			}
		},
		Help: "list protocols for local peer",
	}

	block := &ishell.Cmd{
		Name: "block",
		Help: "send blocks to peers",
	}

	blockFile := &ishell.Cmd{
		Name: "file",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) < 2 {
				c.Println("Peer and file missing")
				return
			}

			toPeer := c.Args[0]
			file := c.Args[1]

			f, err := os.Open(file)
			if err != nil {
				c.Println(err)
				return
			}

			data, err := ioutil.ReadAll(f)
			if err != nil {
				c.Println(err)
				return
			}

			// TODO store filename in a meta
			meta := map[string][]byte{}

			meta["filename"] = []byte(f.Name())

			hsh, n, err := blx.Send(toPeer, data, meta)
			if err != nil {
				c.Println(err)
				return
			}
			c.Printf("Sent block with %d bytes and hash: %s\n", n, hsh)
		},
		Help: "send a file to another peer",
	}

	block.AddCmd(blockFile)

	get := &ishell.Cmd{
		Name: "get",
		Help: "get resource",
	}

	get.AddCmd(getValue)
	get.AddCmd(getProvider)
	get.AddCmd(getBlock)
	// get.AddCmd(getPeer)

	put := &ishell.Cmd{
		Name: "put",
		Help: "put resource",
	}

	put.AddCmd(putValue)
	put.AddCmd(putProvider)
	// put.AddCmd(putPeer)

	list := &ishell.Cmd{
		Name:    "list",
		Aliases: []string{"l", "ls"},
		Help:    "list cached resources",
	}

	list.AddCmd(listValues)
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
