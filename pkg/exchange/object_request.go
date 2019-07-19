package exchange

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema /object-request -type ObjectRequest -in object_request.go -out object_request_generated.go

// ObjectRequest payload for ObjectRequestType
type ObjectRequest struct {
	ObjectHash string            `json:"objectHash:s"`
	Signature  *crypto.Signature `json:"@signature:o"`
}
