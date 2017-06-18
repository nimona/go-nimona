package dht

type ID string

type Peer struct {
	ID      ID
	Address []string
}
