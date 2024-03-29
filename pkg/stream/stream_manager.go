package stream

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Code-Hex/go-generics-cache/policy/simple"

	"nimona.io/pkg/context"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

type (
	manager struct {
		Network     network.Network
		ObjectStore *sqlobjectstore.Store
		// controller cache
		controllers     *simple.Cache[tilde.Digest, Controller]
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
		controllers: simple.NewCache[tilde.Digest, Controller](),
	}
	if network != nil && resolver != nil {
		m.strategy = NewTopographicalSyncStrategy(
			network,
			resolver,
			objectStore,
		)
		go m.strategy.Serve(ctx, m)
	}
	return m, nil
}

func (m *manager) GetOrCreateController(cid tilde.Digest) (Controller, error) {
	c, err := m.GetController(cid)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	return c, nil
}

func (m *manager) GetController(cid tilde.Digest) (Controller, error) {
	// check if controller is already cached and return it
	m.controllersLock.RLock()
	c, found := m.controllers.Get(cid)
	if found {
		m.controllersLock.RUnlock()
		return c, nil
	}
	m.controllersLock.RUnlock()

	// create a new controller
	m.controllersLock.Lock()
	c = NewController(
		cid,
		m.Network,
		m.ObjectStore,
	)

	m.controllers.Set(cid, c)
	m.controllersLock.Unlock()

	// apply the stream to the controller
	r, err := m.ObjectStore.GetByStream(cid)
	if err != nil {
		return c, ErrNotFound
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

	return c, nil
}

func (m *manager) Fetch(
	ctx context.Context,
	ctrl Controller,
	cid tilde.Digest,
) (int, error) {
	return m.strategy.Fetch(ctx, ctrl, cid)
}
