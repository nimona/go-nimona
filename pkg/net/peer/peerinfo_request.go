package peer

//go:generate go run nimona.io/tools/objectify -schema /peer.request -type PeerInfoRequest -in peerinfo_request.go -out peerinfo_request_generated.go

// PeerInfoRequest is a request for a peer info
type PeerInfoRequest struct {
	Keys         []string
	Protocols    []string
	ContentIDs   []string
	ContentTypes []string
}
