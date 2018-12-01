package net

import (
	"nimona.io/go/crypto"
)

// BlockRequest payload for BlockRequestType
type BlockRequest struct {
	RequestID string            `json:"requestID"`
	ID        string            `json:"id"`
	Signature *crypto.Signature `json:"signature"`
	Sender    *crypto.Key       `json:"sender"`
	response  chan interface{}
}
