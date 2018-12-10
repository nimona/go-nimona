package net

import (
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
)

//go:generate go run nimona.io/go/cmd/objectify -schema /block-forward-request -type BlockForwardRequest -out block_forward_request_generated.go

// BlockForwardRequest is the payload for proxied blocks
type BlockForwardRequest struct {
	Recipient string            `json:"recipient"` // address
	FwBlock   *encoding.Object  `json:"fwBlock"`
	Signature *crypto.Signature `json:"@signature"`
}
