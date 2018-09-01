package net

import (
	blocks "github.com/nimona/go-nimona/blocks"
)

func init() {
	blocks.RegisterContentType("blx.response", BlockResponse{})
}

// BlockResponse -
type BlockResponse struct {
	RequestID string `nimona:"requestID" json:"requestID"`
	Block     []byte `nimona:"block" json:"block"`
}
