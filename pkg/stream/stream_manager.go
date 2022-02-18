package stream

import (
	"errors"
	"fmt"
	"sync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

type (
	manager struct {
		Network     network.Network
		ObjectStore *sqlobjectstore.Store
		// controller cache
		controllers     []Controller
		controllersLock sync.RWMutex
		// sync strategy
		strategy SyncStrategy
	}
)

func NewManager(
	ctx context.Context,
	network network.Network,
	resolver resolver.Resolver,
	objectStore *sqlobjectstore.Store,
) (Manager, error) {
	m := &manager{
		Network:     network,
		ObjectStore: objectStore,
	}
	return m, nil
}

func (m *manager) NewController() Controller {
	c := NewController(
		m.Network,
		m.ObjectStore,
	)
	m.controllersLock.Lock()
	m.controllers = append(m.controllers, c)
	m.controllersLock.Unlock()
	return c
}

func (m *manager) GetController(h tilde.Digest) (Controller, error) {
	// check if controller is already cached and return it
	m.controllersLock.RLock()
	for _, controller := range m.controllers {
		if controller.GetStreamRoot() == h {
			m.controllersLock.RUnlock()
			return controller, nil
		}
	}
	m.controllersLock.RUnlock()

	// create a new controller
	c := m.NewController()

	// apply the stream to the controller
	r, err := m.ObjectStore.GetByStream(h)
	if err != nil {
		return nil, fmt.Errorf("error getting stream: %v", err)
	}

	for {
		obj, err := r.Read()
		if err != nil {
			if errors.Is(err, object.ErrReaderDone) {
				break
			}
			return nil, fmt.Errorf("error reading stream: %v", err)
		}
		err = c.Apply(obj)
		if err != nil {
			return nil, fmt.Errorf("error applying object to stream: %v", err)
		}
	}

	m.controllersLock.Lock()
	m.controllers = append(m.controllers, c)
	m.controllersLock.Unlock()

	return c, nil
}

func (m *manager) Fetch(
	ctx context.Context,
	ctrl Controller,
	cid tilde.Digest,
) (int, error) {
	return m.strategy.Fetch(ctx, ctrl, cid)
}
