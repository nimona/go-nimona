package mesh

import (
	"context"
	"errors"
	"time"

	"github.com/nimona/go-nimona/net"
	"go.uber.org/zap"
)

// Mesh is a p2p layer
type Mesh interface {
	Dial(ctx context.Context, peerID, protocol string) (context.Context, net.Conn, error)
}

type mesh struct {
	net      net.Net
	registry Registry
}

func NewMesh(nnet net.Net, registry Registry) (Mesh, error) {
	m := &mesh{
		net:      nnet,
		registry: registry,
	}

	go func() {
		// TODO(geoah) quick hack as sometimes on startup this wasn't triggered fast enough, need to debug why
		time.Sleep(time.Second)
		localPeerInfo := &PeerInfo{
			ID:        registry.GetLocalPeerInfo().ID,
			Protocols: nnet.GetProtocols(),
		}
		registry.PutLocalPeerInfo(localPeerInfo)
	}()

	return m, nil
}

func (m *mesh) Dial(ctx context.Context, peerID, protocol string) (context.Context, net.Conn, error) {
	logger := net.Logger(ctx).Named("mesh")
	logger.Info("mesh.Dial-ing", zap.String("peer_id", peerID), zap.String("protocol", protocol))

	peerInfo, err := m.registry.GetPeerInfo(peerID)
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

	logger.Info("mesh.Dial-failed")
	return nil, nil, errors.New("all addresses failed to dial")
}
