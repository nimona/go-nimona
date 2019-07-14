package exchange

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate $GOBIN/objectify -schema /object-forward-request -type ObjectForwardRequest -in object_forward_request.go -out object_forward_request_generated.go

// ObjectForwardRequest is the payload for proxied objects
type ObjectForwardRequest struct {
	Recipient string            `json:"recipient"` // address
	FwObject  *object.Object    `json:"fwObject"`
	Signature *crypto.Signature `json:"@signature"`
}
