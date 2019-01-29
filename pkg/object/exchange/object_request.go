package exchange

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema /object-request -type ObjectRequest -in object_request.go -out object_request_generated.go

// ObjectRequest payload for ObjectRequestType
type ObjectRequest struct {
	RequestID string            `json:"requestID"`
	ID        string            `json:"id"`
	Signature *crypto.Signature `json:"@signature"`
	Signer    *crypto.Key       `json:"@signer"`
	response  chan *object.Object
}
