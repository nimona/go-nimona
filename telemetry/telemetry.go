package telemetry

import (
	"context"
	"errors"
	"log"
	"os"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
)

type Exchanger interface {
	Send(ctx context.Context, o *encoding.Object, address string) error
	Handle(contentType string, h func(o *encoding.Object) error) (func(), error)
}

const connectionEventType = "nimona.io/telemetry.connection"
const blockEventType = "nimona.io/telemetry.block"

var DefaultClient *metrics

type metrics struct {
	exchange     Exchanger
	colletor     Collector
	localPeer    *crypto.Key
	statsAddress string
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
	statsAddress string) error {

	// create the default client
	DefaultClient = &metrics{
		exchange:     exchange,
		colletor:     DefaultCollector,
		localPeer:    localPeer,
		statsAddress: statsAddress,
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

func (t *metrics) SendEvent(ctx context.Context, event Collectable) error {
	return t.exchange.Send(ctx, event.ToObject(), t.statsAddress)
}

func (t *metrics) handleBlock(o *encoding.Object) error {
	// switch o.GetType() {
	// case *ConnectionEvent:
	// 	t.colletor.Collect(v)
	// case *BlockEvent:
	// 	t.colletor.Collect(v)
	// }

	return nil
}
