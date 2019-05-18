package dht

//go:generate $GOBIN/objectify -schema nimona.io/dht/peerinfo.request -type PeerInfoRequest -in peerinfo_request.go -out peerinfo_request_generated.go

// PeerInfoRequest payload
type PeerInfoRequest struct {
	RequestID   string `json:"requestID,omitempty"`
	Fingerprint string `json:"fingerprint"`
}
