package mesh

import (
	"context"

	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/messenger"
	"github.com/nimona/go-nimona/net"
)

// Mesh is a p2p layer
type Mesh interface {
	Dial(ctx context.Context, address string) (context.Context, net.Conn, error)
	Resolve(ctx context.Context, peerID string) (string, error)
	Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error)
	Publish(ctx context.Context, msg messenger.Message, topics ...string) error
	Subscribe(topics ...string) (chan messenger.Message, error)
	Unsubscribe(chan messenger.Message) error
}

type mesh struct {
	net net.Net
	dht *dht.DHT
}

func NewMesh(nn net.Net, dn *dht.DHT) Mesh {
	return &mesh{
		net: nn,
		dht: dn,
	}
}

func (m *mesh) Dial(ctx context.Context, address string) (context.Context, net.Conn, error) {
	return m.net.DialContext(ctx, address)
}

func (m *mesh) Resolve(ctx context.Context, peerID string) (string, error) {
	ch, err := m.dht.Get(ctx, peerID)
	if err != nil {
		return "", err
	}

	peerAddr := <-ch
	return peerAddr, nil
}

func (m *mesh) Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error) {
	return nil, nil
}

func (m *mesh) Publish(ctx context.Context, msg messenger.Message, topics ...string) error {
	return nil
}

func (m *mesh) Subscribe(topics ...string) (chan messenger.Message, error) {
	return nil, nil
}

func (m *mesh) Unsubscribe(chan messenger.Message) error {
	return nil
}
