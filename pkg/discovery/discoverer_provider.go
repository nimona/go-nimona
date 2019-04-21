package discovery

import (
	"nimona.io/internal/context"
	"nimona.io/pkg/net/peer"
)

// nolint: lll
//go:generate go run github.com/vektra/mockery/cmd/mockery -name Provider -case underscore

// Provider defines the interface for a discoverer provider, eg our DHT
type Provider interface {
	Discover(
		ctx context.Context,
		q *peer.PeerInfoRequest,
	) ([]*peer.PeerInfo, error)
}
