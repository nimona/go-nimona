package net

import (
	blocks "github.com/nimona/go-nimona/blocks"
)

const (
	TypeForwarded = "nimona.forwarded"
)

func init() {
	blocks.RegisterContentType(TypeForwarded, PayloadForwarded{})
}

// PayloadForwarded is the payload for proxied blocks
type PayloadForwarded struct {
	RecipientID string        `json:"recipientID"`
	Block       *blocks.Block `json:"block"`
}
