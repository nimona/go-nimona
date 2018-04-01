package mutation

var (
	PeerProtocolDiscoveredTopic = "peer:protocol:discovered"
	// PeerAddressDiscoveredTopic  = "peer:address:discovered"
	PeerProtocolExpiredTopic = "peer:protocol:expired"
	// PeerAddressExpiredTopic     = "peer:address:expired"
)

// PeerProtocolDiscovered is being published by discovery services
type PeerProtocolDiscovered struct {
	PeerID          string
	ProtocolName    string
	ProtocolAddress string
	Pinned          bool
}

// PeerAddressDiscovered is being published by discovery services
// type PeerAddressDiscovered struct {
// 	PeerID      string
// 	PeerAddress string
// 	Pinned      bool
// }

// PeerProtocolExpired is being published by discovery services and the registry
type PeerProtocolExpired struct {
	PeerID          string
	ProtocolName    string
	ProtocolAddress string
}

// PeerAddressExpired is being published by discovery services and the registry
// type PeerAddressExpired struct {
// 	PeerID      string
// 	PeerAddress string
// }
