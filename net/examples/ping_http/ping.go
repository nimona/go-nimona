package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	nnet "github.com/nimona/go-nimona/net"
	prot "github.com/nimona/go-nimona/net/protocol"
)

// Ping is our example client, it simply sends a PING string and expects a PONG
type Ping struct{}

// Name of our protocol
func (p *Ping) Name() string {
	return "ping"
}

// Negotiate will be called after all the other protocol have been processed
func (p *Ping) Ping(c net.Conn) {
	client, _ := prot.NewHTTPClient(c)
	resp, err := client.Get("http://something/ping")
	if err != nil {
		log.Fatal("get err", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("get read err", err)
	}
	fmt.Println("get", resp.StatusCode, string(body))
}

func (p *Ping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there hon")
}

func (p *Ping) Negotiate(fn nnet.NegotiatorFunc) nnet.NegotiatorFunc {
	return func(ctx context.Context, c nnet.Conn) error {
		return fn(ctx, c)
	}
}

// Handle ping requests
func (p *Ping) Handle(fn nnet.HandlerFunc) nnet.HandlerFunc {
	return func(ctx context.Context, c nnet.Conn) error {
		return prot.NewHTTPServer(c, p)
	}
}
