package network

import (
	"nimona.io/pkg/did"
	"nimona.io/pkg/object"
)

// Envelope -
type Envelope struct {
	Sender  did.DID
	Payload *object.Object
}
