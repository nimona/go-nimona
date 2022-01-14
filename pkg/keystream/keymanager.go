package keystream

import (
	"fmt"
	"sync"

	"github.com/geoah/go-pubsub"
	"nimona.io/pkg/configstore"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

const (
	ErrControllerNotFound = errors.Error("keystream: controller not found")
)

type (
	Manager interface {
		GetController() (Controller, error)
		NewController(*DelegatorSeal) (Controller, error)
		NewDelegationRequest(
			context.Context,
			DelegationRequestVendor,
			Permissions,
		) (*DelegationRequest, chan Controller, error)
		HandleDelegationRequest(
			context.Context,
			*DelegationRequest,
		) error
		WaitForController(context.Context) (Controller, error)
		// WaitForDelegationRequests(context.Context) (chan *DelegationRequest, error)
	}
	manager struct {
		mutex       sync.RWMutex
		network     network.Network
		objectStore *sqlobjectstore.Store
		configStore configstore.Store
		controller  Controller
		topic       *pubsub.Topic[Controller]
	}
)

func NewKeyManager(
	net network.Network,
	objectStore *sqlobjectstore.Store,
	configStore configstore.Store,
) (Manager, error) {
	m := &manager{
		network:     net,
		objectStore: objectStore,
		configStore: configStore,
		topic:       pubsub.NewTopic[Controller](),
	}

	// find controller from config
	controllerHash, err := configStore.Get(configstore.ConfigKeyManagerController)
	if err == nil && controllerHash != "" {
		// load the controller
		reader, err := objectStore.GetByStream(
			tilde.Digest(controllerHash),
		)
		if err != nil {
			return nil, fmt.Errorf("could not load stream: %w", err)
		}

		for {
			streamRoot, err := reader.Read()
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
			m.controller = c
		}
	}

	return m, nil
}

// NewController creates a new controller in the manager's objectstore
func (m *manager) NewController(
	delegatorSeal *DelegatorSeal,
) (Controller, error) {
	// create controller
	c, err := NewController(
		m.network.GetConnectionInfo().Metadata.Owner,
		m.objectStore,
		m.objectStore,
		delegatorSeal,
	)
	if err != nil {
		return nil, err
	}
	// put controller in config
	err = m.configStore.Put(
		configstore.ConfigKeyManagerController,
		string(c.GetKeyStream().Root),
	)
	if err != nil {
		return nil, err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.controller = c
	m.topic.Publish(c)
	return c, nil
}

func (m *manager) GetController() (Controller, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.controller != nil {
		return m.controller, nil
	}
	return nil, ErrControllerNotFound
}

func (m *manager) WaitForController(ctx context.Context) (Controller, error) {
	select {
	case c := <-m.topic.Subscribe():
		return c, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
