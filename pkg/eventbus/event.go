package eventbus

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

type (
	// baseEvent for getting the events to match the event interface
	baseEvent struct{}
	// NetworkAddressAdded is published when the network starts listening
	// to a new address.
	NetworkAddressAdded struct {
		baseEvent
		Address string
	}
	// NetworkAddressRemoved is published when the network stops listening
	// to an existing address.
	NetworkAddressRemoved struct {
		baseEvent
		Address string
	}
	// ObjectPinned is published when an object is pinned.
	ObjectPinned struct {
		baseEvent
		Hash object.Hash
	}
	// ObjectUnpinned is published when an object is unpinned.
	ObjectUnpinned struct {
		baseEvent
		Hash object.Hash
	}
	// PeerConnectionEstablished is published when a connection is
	// established with a peer.
	PeerConnectionEstablished struct {
		baseEvent
		PublicKey crypto.PublicKey
	}
	// RelayAdded is published when a relay has been found and should be used.
	RelayAdded struct {
		baseEvent
		PublicKey crypto.PublicKey
	}
	// RelayRemoved is published when a relay should be removed.
	RelayRemoved struct {
		baseEvent
		PublicKey crypto.PublicKey
	}
)

func (be baseEvent) isLocalEvent() {}
