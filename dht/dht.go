package dht

// DHT ..
type DHT interface {
	Find(ID) (Peer, error)
	Ping(Peer) (Peer, error)
}
