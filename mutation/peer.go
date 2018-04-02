package mutation

var (
	PeerProtocolDiscoveredTopic = "peer:protocol:discovered"
	PeerProtocolExpiredTopic    = "peer:protocol:expired"
)

// PeerProtocolDiscovered is being published by discovery services
type PeerProtocolDiscovered struct {
	PeerID          string
	ProtocolName    string
	ProtocolAddress string
	Pinned          bool
}

// PeerProtocolExpired is being published by discovery services and the registry
type PeerProtocolExpired struct {
	PeerID          string
	ProtocolName    string
	ProtocolAddress string
}
