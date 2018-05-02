package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/net"
	"github.com/nimona/go-nimona/net/protocol"
	"github.com/nimona/go-nimona/wire"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	peerID := os.Getenv("PEER_ID")
	if peerID == "" {
		log.Fatal("Missing PEER_ID")
	}

	bs := []string{}
	port := 0

	if peerID == "bootstrap" {
		port = 26801
	} else {
		bs = append(bs, "tcp:localhost:26801/router/wire")
	}

	ctx := context.Background()
	tcp := net.NewTransportTCP("0.0.0.0", port)

	net := net.New(ctx)
	rtr := protocol.NewRouter()

	reg, _ := mesh.NewRegisty(peerID)
	msh, _ := mesh.NewMesh(net, reg)
	wre, _ := wire.NewWire(msh, reg)
	dht, _ := dht.NewDHT(wre, reg, peerID, true, bs...)

	net.AddProtocols(wre)

	rtr.AddRoute(wre)

	net.AddTransport(tcp, rtr)

	if peerID == "bootstrap" {
		// ds.Put(ctx, "a", "a", map[string]string{})
	}

	shell := ishell.New()
	shell.Printf("Nimona DHT (%s)\n", version)

	// handle get
	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) == 0 {
				c.Println("Missing key")
				return
			}

			key := c.Args[0]
			ctx := context.Background()
			rs, err := dht.Get(ctx, key)
			if err != nil {
				c.Printf("Could not get %s\n", key)
				c.Printf("Error: %s\n", err)
			}
			c.Printf(" - %s", rs)
		},
		Help: "get a value from the dht",
	})

	// handle put
	shell.AddCmd(&ishell.Cmd{
		Name: "put",
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
			if err := dht.Put(ctx, key, val); err != nil {
				c.Printf("Could not get %s\n", key)
				c.Printf("Error: %s\n", err)
			}
		},
		Help: "put a value on the dht",
	})

	// handle providers
	shell.AddCmd(&ishell.Cmd{
		Name: "providers",
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
	})

	// handle values
	shell.AddCmd(&ishell.Cmd{
		Name: "values",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := dht.GetAllValues()
			for key, val := range ps {
				c.Printf("* %s: %s\n", key, val)
			}
		},
		Help: "list all providers stored in our local dht",
	})

	// handle peers
	shell.AddCmd(&ishell.Cmd{
		Name: "peers",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := reg.GetAllPeerInfo()
			for _, peer := range ps {
				c.Println("* " + peer.ID)
				for name, addresses := range peer.Protocols {
					c.Printf("  - %s\n", name)
					for _, address := range addresses {
						c.Printf("     - %s\n", address)
					}
				}
			}
		},
		Help: "list all values stored in our local dht",
	})

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
