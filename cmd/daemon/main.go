package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nimona/go-nimona/blx"
	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/net"
	"github.com/nimona/go-nimona/net/protocol"
	"github.com/nimona/go-nimona/wire"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/abiosoft/ishell.v2"
)

var bootstrapPeerIDs = []string{
	"andromeda.nimona.io",
}

func main() {
	peerID := os.Getenv("PEER_ID")
	if peerID == "" {
		log.Fatal("Missing PEER_ID")
	}

	httpPort := "26880"
	if nhp := os.Getenv("HTTP_PORT"); nhp != "" {
		httpPort = nhp
	}

	bsp := []string{}
	rls := []string{}
	port := 0

	bootstrap := isBootstrap(peerID)

	var tcp net.Transport
	if bootstrap {
		fmt.Println("Starting as bootstrap node")
		port = 26801
	} else {
		bsp = bootstrapPeerIDs
	}

	if skipUPNP, _ := strconv.ParseBool(os.Getenv("SKIP_UPNP")); skipUPNP {
		tcp = net.NewTransportTCP("0.0.0.0", port)
	} else {
		tcp = net.NewTransportTCPWithUPNP("0.0.0.0", port)
	}

	ctx := context.Background()
	net := net.New(ctx)
	rtr := protocol.NewRouter()

	rly := protocol.NewRelayProtocol(net, rls)
	mux := protocol.NewYamux()
	reg, _ := mesh.NewRegisty(peerID)
	msh, _ := mesh.NewMesh(net, reg)
	wre, _ := wire.NewWire(msh, reg)
	dht, _ := dht.NewDHT(wre, reg, peerID, true, bsp...)
	blx, _ := blx.NewBlockExchange(pbs)

	net.AddProtocols(rly)
	net.AddProtocols(mux)
	net.AddProtocols(wre)

	rtr.AddRoute(wre)
	// rtr.AddRoute(rly)

	net.AddTransport(mux, rtr)
	net.AddTransport(tcp, mux, rtr)

	router := gin.Default()
	router.Use(cors.Default())
	local := router.Group("/api/v1/local")
	local.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, reg.GetLocalPeerInfo())
	})
	peers := router.Group("/api/v1/peers")
	peers.GET("/", func(c *gin.Context) {
		peers, err := reg.GetAllPeerInfo()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, peers)
	})
	values := router.Group("/api/v1/values")
	values.GET("/", func(c *gin.Context) {
		values, err := dht.GetAllValues()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, values)
	})
	providers := router.Group("/api/v1/providers")
	providers.GET("/", func(c *gin.Context) {
		providers, err := dht.GetAllProviders()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(http.StatusOK, providers)
	})
	go router.Run(":" + httpPort)

	shell := ishell.New()
	shell.Println("Nimona DHT")

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
				for name, addresses := range peer.Protocols {
					c.Printf("  - %s\n", name)
					for _, address := range addresses {
						c.Printf("     - %s\n", address)
					}
				}
			}
		},
		Help: "list all peers stored in our local dht",
	}

	listLocal := &ishell.Cmd{
		Name: "local",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			peer := reg.GetLocalPeerInfo()
			c.Println("* " + peer.ID)
			for name, addresses := range peer.Protocols {
				c.Printf("  - %s\n", name)
				for _, address := range addresses {
					c.Printf("     - %s\n", address)
				}
			}
		},
		Help: "list protocols for local peer",
	}

	get := &ishell.Cmd{
		Name: "get",
		Help: "get resource",
	}

	get.AddCmd(getValue)
	get.AddCmd(getProvider)
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

	shell.AddCmd(get)
	shell.AddCmd(put)
	shell.AddCmd(list)

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

			blx.Send(toPeer, peerID, f,
				map[string][]byte{})
		},
		Help: "send a file to another peer",
	}

	block.AddCmd(blockFile)
	shell.AddCmd(block)

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

func isBootstrap(peerID string) bool {
	for _, bpi := range bootstrapPeerIDs {
		if bpi == peerID {
			return true
		}
	}
	return false
}
