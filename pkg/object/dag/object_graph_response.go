package dag

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema /object-graph-response -type ObjectGraphResponse -in object_graph_response.go -out object_graph_response_generated.go

// ObjectGraphResponse is the payload for proxied objects
type ObjectGraphResponse struct {
	ObjectHashes []string          `json:"objectHashes:as"`
	Signature    *crypto.Signature `json:"@signature:o"`
}
