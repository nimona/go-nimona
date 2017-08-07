package dht

type RoutingTable interface {
	Save(Peer) error
	Remove(Peer) error
	Get(ID) (Peer, error)
	GetPeerIDs() ([]ID, error)
}
