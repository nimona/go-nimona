package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	uuid "github.com/google/uuid"
	logrus "github.com/sirupsen/logrus"
	ishell "gopkg.in/abiosoft/ishell.v2"

	fabric "github.com/nimona/go-nimona-fabric"
	dht "github.com/nimona/go-nimona-fabric/dht"
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

	absp := map[string][]string{
		"bootstrap": []string{
			"tcp:192.168.0.10:26800",
		},
	}
	pid := uuid.New().String()
	bsp := map[string][]string{}
	if cpid := os.Getenv("PEER_ID"); cpid != "" {
		pid = cpid
		for prID, pr := range absp {
			if cpid == prID {
				logrus.Warnf("Skipping bootstrap peer %s", cpid)
				continue
			}
			bsp[prID] = pr
		}
	} else {
		bsp = absp
	}

	ctx := context.Background()
	nn := fabric.New(ctx)

	dn, err := dht.NewDHT(bsp, pid, nn)
	if err != nil {
		logrus.WithError(err).Fatalf("Could not get dht")
	}

	tcp := fabric.NewTransportTCP("0.0.0.0", port)
	nn.AddTransport(tcp, []fabric.Protocol{dn})
	fmt.Println("Addresses: ", nn.GetAddresses())

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
