package exchange

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema /object-response -type ObjectResponse -in object_response.go -out object_response_generated.go

// ObjectResponse -
type ObjectResponse struct {
	RequestID       string            `json:"requestID"`
	RequestedObject *object.Object    `json:"requestedObject"`
	Sender          *crypto.Key       `json:"sender"`
	Signature       *crypto.Signature `json:"@signature"`
}
