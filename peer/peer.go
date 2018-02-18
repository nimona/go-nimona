package peer

type Peer struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
}

func NewPeer(id string) *Peer {
	return &Peer{
		ID: id,
	}
}
