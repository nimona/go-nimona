package dag

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema /object-graph-request -type ObjectGraphRequest -in object_graph_request.go -out object_graph_request_generated.go

// ObjectGraphRequest is the payload for proxied objects
type ObjectGraphRequest struct {
	Selector  []string          `json:"selector:as"`
	Signature *crypto.Signature `json:"@signature:o"`
}
