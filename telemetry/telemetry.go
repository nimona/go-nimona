package telemetry

import (
	"context"

	blocks "github.com/nimona/go-nimona/blocks"
)

type Exchanger interface {
	Send(ctx context.Context, o interface{}, recipient *blocks.Key,
		opts ...blocks.MarshalOption) error
	Handle(contentType string, h func(o interface{}) error) error
}

type Metrics struct {
	exchange      Exchanger
	colletor      Collector
	localPeer     *blocks.Key
	bootstrapPeer *blocks.Key
}

const connectionEventType = "nimona.telemetry.connection"
const blockEventType = "nimona.telemetry.block"

func NewTelemetry(exchange Exchanger, collector Collector,
	localPeer *blocks.Key, bootstrapPeer *blocks.Key) (*Metrics, error) {
	// Register the two basic types
	blocks.RegisterContentType(connectionEventType, ConnectionEvent{})
	blocks.RegisterContentType(blockEventType, BlockEvent{})
	m := &Metrics{
		exchange:      exchange,
		colletor:      collector,
		localPeer:     localPeer,
		bootstrapPeer: bootstrapPeer,
	}

	// Register handler only on server
	// check the env var
	if collector != nil {
		exchange.Handle(connectionEventType, m.handleBlock)
		exchange.Handle(blockEventType, m.handleBlock)
	}

	return m, nil
}

func (t *Metrics) SendEvent(ctx context.Context,
	event Collectable) error {
	return t.exchange.Send(ctx,
		event, t.bootstrapPeer, blocks.SignWith(t.localPeer))
}

func (t *Metrics) handleBlock(payload interface{}) error {
	switch v := payload.(type) {
	case *ConnectionEvent:
		t.colletor.Collect(v)
	case *BlockEvent:
		t.colletor.Collect(v)
	}

	return nil
}
