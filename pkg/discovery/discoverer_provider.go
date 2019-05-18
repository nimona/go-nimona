package discovery

import (
	"nimona.io/internal/context"
	"nimona.io/pkg/net/peer"
)

// nolint: lll
//go:generate go run github.com/vektra/mockery/cmd/mockery -name Provider -case underscore

// Provider defines the interface for a discoverer provider, eg our DHT
type Provider interface {
	FindByFingerprint(
		ctx context.Context,
		fingerprint string,
		opts ...Option,
	) ([]*peer.PeerInfo, error)
	FindByContent(
		ctx context.Context,
		contentHash string,
		opts ...Option,
	) ([]*peer.PeerInfo, error)
}
