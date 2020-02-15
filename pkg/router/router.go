package router

import (
	"crypto"

	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

type (
	Envelope struct {
		Sender  crypto.PublicKey
		Payload object.Object
	}
	Router interface {
		Send(*peer.Peer, object.Object) error
		Handle() <-chan Envelope
	}
)

func New() Router {
	return nil
}
