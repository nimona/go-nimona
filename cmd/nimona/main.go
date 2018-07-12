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
	"github.com/nimona/go-nimona/net"

	"gopkg.in/abiosoft/ishell.v2"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var bootstrapPeerInfos = []net.PeerInfo{
	net.PeerInfo{
		ID: "P0x3082010a028201010098156fb69d71d06dfd27f1d41541c25e1687e909ff4fd9fe07b7e3041079e19e5ee79b95264ff4eb9c43c6573dd6a83cacb80515801163736d06d17d0cea7a5f2e76992dd4bcdb37b61e01cbcaae9f53c04699f3b95cbef5628d7a9520e244980a0299dcddb458be4fc1bb0e8aed81b890133e7cd28a88a3ba61edada4917af024b1a2eb3ce6a41e7c08ea3a39fecfc1d82fb126189bf5f20e134dbafa03ba6313faca2ea11a30a106039858a37d93e327b3c98b01ba9a46df9bc0d2147edf3cfbfec8641a5eea3448ad30fb34405c85c867f655bb72e2897fd5f5b74c84459ae43200adaa1b4c39e9434587835ed4ad8fc01b1e7785fc8d0f15ee76e60083e10203010001",
		Addresses: []string{
			"tcp:andromeda.nimona.io:21013",
		},
		PublicKey: []byte{
			48, 130, 1, 10, 2, 130, 1, 1, 0, 152, 21, 111, 182, 157, 113,
			208, 109, 253, 39, 241, 212, 21, 65, 194, 94, 22, 135, 233, 9,
			255, 79, 217, 254, 7, 183, 227, 4, 16, 121, 225, 158, 94, 231,
			155, 149, 38, 79, 244, 235, 156, 67, 198, 87, 61, 214, 168, 60,
			172, 184, 5, 21, 128, 17, 99, 115, 109, 6, 209, 125, 12, 234,
			122, 95, 46, 118, 153, 45, 212, 188, 219, 55, 182, 30, 1, 203,
			202, 174, 159, 83, 192, 70, 153, 243, 185, 92, 190, 245, 98,
			141, 122, 149, 32, 226, 68, 152, 10, 2, 153, 220, 221, 180, 88,
			190, 79, 193, 187, 14, 138, 237, 129, 184, 144, 19, 62, 124,
			210, 138, 136, 163, 186, 97, 237, 173, 164, 145, 122, 240, 36,
			177, 162, 235, 60, 230, 164, 30, 124, 8, 234, 58, 57, 254, 207,
			193, 216, 47, 177, 38, 24, 155, 245, 242, 14, 19, 77, 186, 250,
			3, 186, 99, 19, 250, 202, 46, 161, 26, 48, 161, 6, 3, 152, 88,
			163, 125, 147, 227, 39, 179, 201, 139, 1, 186, 154, 70, 223,
			155, 192, 210, 20, 126, 223, 60, 251, 254, 200, 100, 26, 94,
			234, 52, 72, 173, 48, 251, 52, 64, 92, 133, 200, 103, 246, 85,
			187, 114, 226, 137, 127, 213, 245, 183, 76, 132, 69, 154, 228,
			50, 0, 173, 170, 27, 76, 57, 233, 67, 69, 135, 131, 94, 212, 173,
			143, 192, 27, 30, 119, 133, 252, 141, 15, 21, 238, 118, 230, 0,
			131, 225, 2, 3, 1, 0, 1,
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

	keyPath := path.Join(configPath, "net.json")

	port, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 32)

	reg := net.NewAddressBook()
	spi, err := reg.LoadOrCreateLocalPeerInfo(keyPath)
	if err != nil {
		log.Fatal("could not load key", err)
	}
	if err := reg.PutLocalPeerInfo(spi); err != nil {
		log.Fatal("could not put local peer")
	}

	for _, peerInfo := range bootstrapPeerInfos {
		reg.PutPeerInfo(&peerInfo)
	}

	storagePath := path.Join(configPath, "storage")

	n, _ := net.NewWire(reg)
	dht, _ := dht.NewDHT(n, reg)
	dpr := blx.NewDiskStorage(storagePath)
	blx, _ := blx.NewBlockExchange(n, dpr)

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

	n.Listen(fmt.Sprintf("0.0.0.0:%d", port))

	n.HandleExtensionEvents("msg", func(event *net.Message) error {
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
			message, err := net.NewMessage("msg.msg", to, msg)
			if err != nil {
				c.Println("Could not create message", err)
				return
			}
			if err := n.Send(ctx, message); err != nil {
				c.Println("Could not send message", err)
				return
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
