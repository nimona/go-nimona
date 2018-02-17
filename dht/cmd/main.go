package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	uuid "github.com/google/uuid"
	logrus "github.com/sirupsen/logrus"

	dht "github.com/nimona/go-nimona-kad-dht"
	net "github.com/nimona/go-nimona-net"
)

func main() {
	ll := logrus.ErrorLevel
	if ell, err := logrus.ParseLevel(os.Getenv("DEBUG")); err == nil {
		ll = ell
	}
	logrus.SetLevel(ll)

	port := 0
	eprt := os.Getenv("PORT")
	if eprt != "" {
		port, _ = strconv.Atoi(eprt)
	}

	if port == 0 {
		port = net.GetPort()
	}
	addrs, _ := net.GetAddresses(port)
	addrs = append(addrs, fmt.Sprintf("tcp4:0.0.0.0:%d", port))

	absp := []net.Peer{
		{
			ID: "bootstrap.nimona.io",
			Addresses: []string{
				"tcp4:bootstrap.nimona.io:26800",
			},
		},
	}
	pid := uuid.New().String()
	bsp := []net.Peer{}
	if cpid := os.Getenv("PEER_ID"); cpid != "" {
		pid = cpid
		for _, pr := range absp {
			if cpid == pr.ID {
				logrus.Warnf("Skipping bootstrap peer %s", cpid)
				continue
			}
			bsp = append(bsp, pr)
		}
	} else {
		bsp = absp
	}

	pr := &net.Peer{
		ID:        pid,
		Addresses: addrs,
	}

	nn, err := net.NewNetwork(pr, port)
	if err != nil {
		logrus.WithError(err).Fatalf("Could not get network")
	}

	dn, err := dht.NewDHTNode(bsp, *pr, nn)
	if err != nil {
		logrus.WithError(err).Fatalf("Could not get dht")
	}

	shell := ishell.New()
	shell.Println("Nimona DHT")

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
			rs, err := dn.Get(ctx, key)
			if err != nil {
				c.Printf("Could not get %s\n", key)
				c.Printf("Error: %s\n", err)
			}
			c.ProgressBar().Indeterminate(true)
			c.ProgressBar().Start()
			for rv := range rs {
				c.Println("  - " + rv)
			}
			c.ProgressBar().Stop()
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
			if err := dn.Put(ctx, key, val); err != nil {
				c.Printf("Could not get %s\n", key)
				c.Printf("Error: %s\n", err)
			}
		},
		Help: "put a value on the dht",
	})

	// handle put
	shell.AddCmd(&ishell.Cmd{
		Name: "list",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := dn.GetLocalPairs()
			for key, vals := range ps {
				c.Println("* " + key)
				for _, val := range vals {
					c.Println("  - " + val.Value)
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
