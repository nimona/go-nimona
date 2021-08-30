package keystream

import (
	"sync"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

const (
	ErrControllerNotFound = errors.Error("keystream: controller not found")
)

type (
	Manager interface {
		AddController(Controller)
		GetController(tilde.Digest) (Controller, error)
		ListControllers() []Controller
		NewController(*DelegatorSeal) (Controller, error)
	}
	manager struct {
		mutex       sync.RWMutex
		network     network.Network
		objectStore *sqlobjectstore.Store
		controllers map[tilde.Digest]Controller
	}
)

func NewKeyManager(
	net network.Network,
	objectStore *sqlobjectstore.Store,
) (Manager, error) {
	// load all our controllers from the objectstore
	cs := map[tilde.Digest]Controller{}
	streamRootsReader, err := objectStore.Filter(
		sqlobjectstore.FilterByOwner(net.GetConnectionInfo().Metadata.Owner),
		sqlobjectstore.FilterByObjectType("keri.Inception/v0"),
	)
	if err != nil && !errors.Is(err, objectstore.ErrNotFound) {
		return nil, err
	}

	if streamRootsReader != nil {
		for {
			streamRoot, err := streamRootsReader.Read()
			if err != nil {
				if errors.Is(err, object.ErrReaderDone) {
					break
				}
				return nil, err
			}
			c, err := RestoreController(
				streamRoot.Hash(),
				objectStore,
				objectStore,
			)
			if err != nil {
				return nil, err
			}
			cs[streamRoot.Hash()] = c
		}
	}

	return &manager{
		network:     net,
		objectStore: objectStore,
		controllers: cs,
	}, nil
}

// NewController creates a new controller in the manager's objectstore
func (m *manager) NewController(
	delegatorSeal *DelegatorSeal,
) (Controller, error) {
	c, err := NewController(
		m.network.GetConnectionInfo().Metadata.Owner,
		m.objectStore,
		m.objectStore,
		delegatorSeal,
	)
	if err != nil {
		return nil, err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.controllers[c.GetKeyStream().Root] = c
	return c, nil
}

// ListControllers returns a list of all controllers
func (m *manager) ListControllers() []Controller {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	ds := []Controller{}
	for _, d := range m.controllers {
		ds = append(ds, d)
	}
	return ds
}

func (m *manager) GetController(d tilde.Digest) (Controller, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	c, ok := m.controllers[d]
	if !ok {
		return nil, ErrControllerNotFound
	}
	return c, nil
}

func (m *manager) AddController(c Controller) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.controllers[c.GetKeyStream().Root] = c
}
