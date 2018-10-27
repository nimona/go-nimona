package dht // import "nimona.io/go/dht"

import "errors"

var (
	ErrPeerAlreadyExists = errors.New("Peer already exists")
	ErrPeerNotFound      = errors.New("Peer not found")
)
