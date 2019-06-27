package discovery

import (
	"nimona.io/internal/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

// nolint: lll
//go:generate $GOBIN/mockery -name Provider -case underscore

// Provider defines the interface for a discoverer provider, eg our DHT
type Provider interface {
	FindByFingerprint(
		ctx context.Context,
		fingerprint crypto.Fingerprint,
		opts ...Option,
	) ([]*peer.PeerInfo, error)
	FindByContent(
		ctx context.Context,
		contentHash string,
		opts ...Option,
	) ([]*peer.PeerInfo, error)
}
