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

	"github.com/nimona/go-nimona/telemetry"

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

var bootstrapPeerInfoEnvelopes = []*net.Envelope{
	&net.Envelope{
		Type: "peer.info",
		Headers: net.Headers{
			Signer: "01x035de8adad618206455f6b7c2ca4fd943faabcba12ae6fea9d6204760d4d6216ff",
		},
		Payload: net.PeerInfoPayload{
			Addresses: []string{
				"tcp:andromeda.nimona.io:21013",
			},
		},
		Signature: []byte{
			63, 75, 91, 132, 252, 236, 211, 254, 9, 245, 64, 255, 216, 226,
			222, 153, 41, 203, 66, 233, 19, 218, 225, 212, 133, 166, 128,
			93, 115, 28, 85, 1, 87, 219, 10, 223, 126, 26, 134, 96, 56, 86,
			223, 238, 113, 120, 165, 25, 103, 185, 231, 232, 204, 227, 48,
			122, 80, 185, 205, 86, 0, 110, 94, 59,
		},
	},
}

// Hello payload
type Hello struct {
	Body string
}

func init() {
	telemetry.SetupKeenCollector()
	net.RegisterContentType("demo.hello", Hello{})
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

	port, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 32)

	reg := net.NewAddressBook()
	spi, err := reg.LoadOrCreateLocalPeerInfo(configPath)
	if err != nil {
		log.Fatal("could not load key", err)
	}
	if err := reg.PutLocalPeerInfo(spi); err != nil {
		log.Fatal("could not put local peer")
	}

	for _, peerInfoEnvelope := range bootstrapPeerInfoEnvelopes {
		reg.PutPeerInfoFromEnvelope(peerInfoEnvelope)
	}

	// mmmm := &net.Envelope{
	// 	Type: "peer.info",
	// 	Payload: net.PeerInfoPayload{
	// 		Addresses: []string{
	// 			"tcp:andromeda.nimona.io:21013",
	// 		},
	// 	},
	// }
	// mmmm.Sign(spi)
	// bbb, _ := json.MarshalIndent(mmmm, "", "  ")
	// fmt.Println(string(bbb))
	// fmt.Println(mmmm.Signature)
	// os.Exit(1)

	storagePath := path.Join(configPath, "storage")

	n, _ := net.NewMessenger(reg)
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

	n.Listen(context.Background(), fmt.Sprintf("0.0.0.0:%d", port))

	n.Handle("demo", func(envelope *net.Envelope) error {
		payload := envelope.Payload.(Hello)
		fmt.Printf("___ Got envelope %s\n", payload.Body)
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
				c.Println("Missing peer id or envelope")
				return
			}
			ctx := context.Background()
			msg := strings.Join(c.Args[1:], " ")
			to := []string{c.Args[0]}
			envelope := net.NewEnvelope("demo.hello", to, &Hello{msg})
			if err := n.Send(ctx, envelope); err != nil {
				c.Println("Could not send envelope", err)
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
