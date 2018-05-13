package mesh

type PeerInfo struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
}
