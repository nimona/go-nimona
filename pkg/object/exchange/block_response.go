package exchange

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema /block-response -type BlockResponse -in block_response.go -out block_response_generated.go

// BlockResponse -
type BlockResponse struct {
	RequestID      string            `json:"requestID"`
	RequestedBlock *object.Object  `json:"requestedBlock"`
	Sender         *crypto.Key       `json:"sender"`
	Signature      *crypto.Signature `json:"@signature"`
}
