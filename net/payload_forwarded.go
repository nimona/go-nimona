package net

import (
	blocks "github.com/nimona/go-nimona/blocks"
)

const (
	// ForwardRequestType is the type of ForwardRequest Blocks
	ForwardRequestType = "nimona.forwarded"
)

func init() {
	blocks.RegisterContentType(ForwardRequestType, ForwardRequest{})
}

// ForwardRequest is the payload for proxied blocks
type ForwardRequest struct {
	Recipient *blocks.Key `json:"recipient"`
	Block     []byte      `json:"block"`
}
