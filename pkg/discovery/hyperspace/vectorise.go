package hyperspace

import (
	"github.com/james-bowman/sparse"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

// Vectorise returns a sparse vector from a PeerInfoRequest
func Vectorise(q *peer.PeerInfoRequest) *sparse.Vector {
	i := []int{}
	if len(q.Keys) > 0 {
		for _, fingerprint := range q.Keys {
			i = append(i, HashChunked("pk", []byte(fingerprint))...)
		}
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
		Keys:         []crypto.Fingerprint{},
		Protocols:    p.Protocols,
		ContentIDs:   p.ContentIDs,
		ContentTypes: p.ContentTypes,
	}
	keys := crypto.GetSignatureKeys(p.Signature)
	for _, key := range keys {
		q.Keys = append(q.Keys, key.Fingerprint())
	}
	return q
}
