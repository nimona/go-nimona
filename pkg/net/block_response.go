package net

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/encoding"
)

//go:generate go run nimona.io/tools/objectify -schema /block-response -type BlockResponse -out block_response_generated.go

// BlockResponse -
type BlockResponse struct {
	RequestID      string            `json:"requestID"`
	RequestedBlock *encoding.Object  `json:"requestedBlock"`
	Sender         *crypto.Key       `json:"sender"`
	Signature      *crypto.Signature `json:"@signature"`
}
