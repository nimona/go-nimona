package exchange

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

// Envelope -
type Envelope struct {
	RequestID string
	Sender    *crypto.PublicKey
	Payload   *object.Object
}
