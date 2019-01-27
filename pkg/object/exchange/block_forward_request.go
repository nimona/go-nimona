package exchange

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema /block-forward-request -type BlockForwardRequest -in block_forward_request.go -out block_forward_request_generated.go

// BlockForwardRequest is the payload for proxied blocks
type BlockForwardRequest struct {
	Recipient string            `json:"recipient"` // address
	FwBlock   *object.Object    `json:"fwBlock"`
	Signature *crypto.Signature `json:"@signature"`
}
