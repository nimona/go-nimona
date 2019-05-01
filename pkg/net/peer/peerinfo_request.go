package peer

//go:generate go run nimona.io/tools/objectify -schema /peer.request -type PeerInfoRequest -in peerinfo_request.go -out peerinfo_request_generated.go

// PeerInfoRequest is a request for a peer info
type PeerInfoRequest struct {
	SignerKeyHash string   `json:"signer"`
	Protocols     []string `json:"protocols"`
	ContentIDs    []string `json:"contentIDs"`
	ContentTypes  []string `json:"contentTypes"`
}
