package net

import (
	"nimona.io/pkg/crypto"
)

//go:generate go run nimona.io/tools/objectify -schema /block-request -type BlockRequest -out block_request_generated.go

// BlockRequest payload for BlockRequestType
type BlockRequest struct {
	RequestID string            `json:"requestID"`
	ID        string            `json:"id"`
	Signature *crypto.Signature `json:"signature"`
	Sender    *crypto.Key       `json:"sender"`
	response  chan interface{}
}
