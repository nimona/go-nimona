package stream

import (
	"nimona.io/pkg/network"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

type (
	manager struct {
		Network     network.Network
		ObjectStore *sqlobjectstore.Store
	}
)

func NewManager(
	network network.Network,
	objectStore *sqlobjectstore.Store,
) (Manager, error) {
	m := &manager{
		Network:     network,
		ObjectStore: objectStore,
	}
	return m, nil
}

func (m *manager) NewStreamController() Controller {
	return NewController(
		m.Network,
		m.ObjectStore,
	)
}

func (m *manager) GetStreamController(h tilde.Digest) Controller {
	c := &controller{}
	return c
}
