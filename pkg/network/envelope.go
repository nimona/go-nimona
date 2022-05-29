package network

import (
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

// Envelope -
type Envelope struct {
	Sender  peer.ID
	Payload *object.Object
}
