package net

import (
	blocks "github.com/nimona/go-nimona/blocks"
)

const (
	// BlockRequestType is the type of a BlockRequest Block
	BlockRequestType = "blx.request-block"
)

func init() {
	blocks.RegisterContentType(BlockRequestType, BlockRequest{})
}

// BlockRequest payload for BlockRequestType
type BlockRequest struct {
	RequestID string `json:"requestID"`
	ID        string `json:"id"`
	response  chan *blocks.Block
}
