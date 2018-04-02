package mesh

import (
	"context"
	"errors"
	"time"

	"github.com/nimona/go-nimona/mutation"
	"github.com/nimona/go-nimona/net"
	"go.uber.org/zap"
)

// Mesh is a p2p layer
type Mesh interface {
	PubSub
	Registry
	Dial(ctx context.Context, peerID, protocol string) (context.Context, net.Conn, error)
}

type mesh struct {
	net      net.Net
	registry Registry
	pubsub   PubSub
}

func NewMesh(nnet net.Net, ps PubSub, registry Registry) (Mesh, error) {
	m := &mesh{
		net:      nnet,
		registry: registry,
		pubsub:   ps,
	}

	go func() {
		// TODO(geoah) quick hack as sometimes on startup this wasn't triggered fast enough, need to debug why
		time.Sleep(time.Second)
		localPeerID := registry.GetLocalPeerInfo().ID
		for protocolName, protocols := range nnet.GetProtocols() {
			for _, protocolAddress := range protocols {
				msg := mutation.PeerProtocolDiscovered{
					PeerID:          localPeerID,
					ProtocolName:    protocolName,
					ProtocolAddress: protocolAddress,
					Pinned:          true,
				}
				ps.Publish(msg, mutation.PeerProtocolDiscoveredTopic)
			}
		}
	}()

	return m, nil
}

func (m *mesh) Dial(ctx context.Context, peerID, protocol string) (context.Context, net.Conn, error) {
	logger := net.Logger(ctx).Named("mesh")
	logger.Info("mesh.Dial-ing", zap.String("peer_id", peerID), zap.String("protocol", protocol))

	peerInfo, err := m.registry.GetPeerInfo(ctx, peerID)
	if err != nil {
		return nil, nil, err
	}

	if len(peerInfo.Protocols) == 0 {
		return nil, nil, errors.New("no protocols available")
	}

	for _, protocolAddress := range peerInfo.Protocols[protocol] {
		nctx, ncon, nerr := m.net.DialContext(ctx, protocolAddress)
		if nerr != nil {
			continue
		}
		logger.Info("mesh.Dial-ed", zap.String("address", protocolAddress))
		return nctx, ncon, nil
	}

	logger.Warn("mesh.Dial-failed")
	return nil, nil, errors.New("all addresses failed to dial")
}

func (m *mesh) Publish(event interface{}, topic string) error {
	return m.pubsub.Publish(event, topic)
}

func (m *mesh) Subscribe(topic string) (chan interface{}, error) {
	return m.pubsub.Subscribe(topic)
}

func (m *mesh) Unsubscribe(ch chan interface{}) error {
	return m.pubsub.Unsubscribe(ch)
}

func (m *mesh) GetLocalPeerInfo() PeerInfo {
	return m.registry.GetLocalPeerInfo()
}

func (m *mesh) GetPeerInfo(ctx context.Context, peerID string) (PeerInfo, error) {
	return m.registry.GetPeerInfo(ctx, peerID)
}

func (m *mesh) GetAllPeerInfo(ctx context.Context) ([]PeerInfo, error) {
	return m.registry.GetAllPeerInfo(ctx)
}
