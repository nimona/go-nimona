package hyperspace

import (
	"github.com/james-bowman/sparse"

	"nimona.io/pkg/net/peer"
)

// Vectorise returns a sparse vector from a PeerInfoRequest
func Vectorise(q *peer.PeerInfoRequest) *sparse.Vector {
	i := []int{}
	if q.SignerKeyHash != "" {
		i = append(i, HashChunked("sk", []byte(q.SignerKeyHash))...)
	}
	if len(q.Protocols) > 0 {
		for _, protocol := range q.Protocols {
			i = append(i, HashChunked("p", []byte(protocol))...)
		}
	}
	if len(q.ContentTypes) > 0 {
		for _, contentType := range q.ContentTypes {
			i = append(i, HashChunked("c", []byte(contentType))...)
		}
	}
	if len(q.ContentIDs) > 0 {
		for _, contentID := range q.ContentIDs {
			i = append(i, HashChunked("cid", []byte(contentID))...)
		}
	}
	d := []float64{}
	for range i {
		d = append(d, 1)
	}
	v := sparse.NewVector(int(scaledMax), i, d)
	return v
}

func getPeerInfoRequest(p *peer.PeerInfo) *peer.PeerInfoRequest {
	q := &peer.PeerInfoRequest{
		Protocols:    p.Protocols,
		ContentIDs:   p.ContentIDs,
		ContentTypes: p.ContentTypes,
	}
	if p.SignerKey != nil {
		q.SignerKeyHash = p.SignerKey.Hash
	}
	return q
}
