package dht

type RoutingTable interface {
	Add(Peer) error
	Remove(Peer) error
	Update(Peer) error
	Get(ID) (Peer, error)
	GetPeerIDs() ([]ID, error)
}
