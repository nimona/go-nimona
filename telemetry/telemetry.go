package telemetry

import (
	"context"
	"errors"
	"log"
	"os"

	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

type Exchanger interface {
	Send(ctx context.Context, o blocks.Typed, recipient *crypto.Key,
		opts ...blocks.PackOption) error
	Handle(contentType string, h func(o blocks.Typed) error) (func(), error)
}

const connectionEventType = "nimona.telemetry.connection"
const blockEventType = "nimona.telemetry.block"

var DefaultClient *metrics

type metrics struct {
	exchange      Exchanger
	colletor      Collector
	localPeer     *crypto.Key
	bootstrapPeer *crypto.Key
}

func init() {
	// Check if the telemetry flags are set and start the collector
	if os.Getenv("TELEMETRY") == "server" {
		user := os.Getenv("TELEMETRY_SERVER_USER")
		pass := os.Getenv("TELEMETRY_SERVER_PASS")
		addr := os.Getenv("TELEMETRY_SERVER_ADDRESS")

		if user != "" || addr != "" {
			col, err := NewInfluxCollector(user, pass, addr)
			if err != nil {
				log.Println("Failed to connect to inlfux", err)
			}

			DefaultCollector = col
		}
	}
}

func NewTelemetry(exchange Exchanger, localPeer *crypto.Key,
	bootstrapPeer *crypto.Key) error {
	// Register the two basic types
	blocks.RegisterContentType(&ConnectionEvent{})
	blocks.RegisterContentType(&BlockEvent{})

	// create the default client
	DefaultClient = &metrics{
		exchange:      exchange,
		colletor:      DefaultCollector,
		localPeer:     localPeer,
		bootstrapPeer: bootstrapPeer,
	}

	// Register handler only on server
	// check the env var
	// TODO is this actually nil at some point?
	if DefaultClient.colletor != nil {
		exchange.Handle(connectionEventType, DefaultClient.handleBlock)
		exchange.Handle(blockEventType, DefaultClient.handleBlock)
	}

	return nil
}

func SendEvent(ctx context.Context, event Collectable) error {
	if DefaultClient == nil {
		return errors.New("no default client")
	}
	return DefaultClient.SendEvent(ctx, event)
}

func (t *metrics) SendEvent(ctx context.Context,
	event Collectable) error {
	return t.exchange.Send(ctx,
		event, t.bootstrapPeer, blocks.SignWith(t.localPeer))
}

func (t *metrics) handleBlock(payload blocks.Typed) error {
	switch v := payload.(type) {
	case *ConnectionEvent:
		t.colletor.Collect(v)
	case *BlockEvent:
		t.colletor.Collect(v)
	}

	return nil
}
