package backlog

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

type (
	// AckFunc allows the handler of a popped object to acknowledge it has been
	// handled and thus can be removed from the backlog.
	AckFunc func()
	// Backlog keeps track of objects than need to be sent to each recipient.
	Backlog interface {
		Push(object.Object, ...crypto.PublicKey) error
		Pop(crypto.PublicKey) (object.Object, AckFunc, error)
		// Peek(*crypto.Key()) (object.Object, error)
	}
)
