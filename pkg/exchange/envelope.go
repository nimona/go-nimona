package exchange

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
)

// Envelope -
type Envelope struct {
	RequestID string
	Sender    *crypto.PublicKey
	Payload   object.Object

	conn *net.Connection
}

func (e *Envelope) Respond(o object.Object) error {
	if e.RequestID != "" {
		o.Set(ObjectRequestID, e.RequestID)
	}
	return net.Write(o, e.conn)
}
