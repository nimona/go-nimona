package mesh

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/nimona/go-nimona/mutation"
)

var (
	ErrNotKnown = errors.New("not found")
)

var (
	peerInfoExpireAfter = time.Hour * 1
)

type Registry interface {
	GetLocalPeerInfo() PeerInfo
	GetPeerInfo(ctx context.Context, peerID string) (PeerInfo, error)
	// Resolve(ctx context.Context, peerID string) (string, error)
	// Discover(ctx context.Context, peerID, protocol string) ([]net.Address, error)
}

func NewRegisty(peerID string, ps PubSub) (Registry, error) {
	incoming, err := ps.Subscribe("peer:.*")
	if err != nil {
		return nil, err
	}

	reg := &registry{
		localPeer: PeerInfo{
			ID:        peerID,
			Protocols: map[string][]string{},
		},
		incoming:  incoming,
		protocols: sync.Map{},
	}

	go func() {
		for event := range reg.incoming {
			switch mut := event.(type) {
			case mutation.PeerProtocolDiscovered:
				protocol := &peerInfoProtocol{
					PeerID:      mut.PeerID,
					Name:        mut.ProtocolName,
					Address:     mut.ProtocolAddress,
					LastUpdated: time.Now(),
					Pinned:      mut.Pinned,
				}
				reg.protocols.Store(protocol.Hash(), protocol)
			}
		}
	}()

	return reg, nil
}

type registry struct {
	protocols sync.Map
	localPeer PeerInfo
	incoming  chan interface{}
	// protocols []*peerInfoProtocol
}

func (reg *registry) GetLocalPeerInfo() PeerInfo {
	return reg.localPeer
}

func (reg *registry) GetPeerInfo(ctx context.Context, peerID string) (PeerInfo, error) {
	peerInfo := PeerInfo{
		ID:        peerID,
		Protocols: map[string][]string{},
	}

	protocols := []*peerInfoProtocol{}
	reg.protocols.Range(func(k, v interface{}) bool {
		protocol := v.(*peerInfoProtocol)
		if strings.HasPrefix(k.(string), peerID) {
			protocols = append(protocols, protocol)
		}
		return true
	})

	for _, protocolInfo := range protocols {
		if protocolInfo.PeerID != peerID {
			continue
		}

		name := protocolInfo.Name
		if _, ok := peerInfo.Protocols[name]; !ok {
			peerInfo.Protocols[name] = []string{}
		}

		expired := protocolInfo.LastUpdated.Add(peerInfoExpireAfter).Before(time.Now())
		if protocolInfo.Pinned == false && expired {
			continue
		}

		peerInfo.Protocols[name] = append(peerInfo.Protocols[name], protocolInfo.Address)
	}

	return peerInfo, nil
}
