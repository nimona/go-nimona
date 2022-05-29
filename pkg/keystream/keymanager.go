package keystream

import (
	"fmt"
	"sync"

	"github.com/geoah/go-pubsub"

	"nimona.io/pkg/configstore"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/network"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
	"nimona.io/pkg/tilde"
)

const (
	ErrControllerNotFound = errors.Error("keystream: controller not found")
)

//go:generate mockgen -destination=../keystreammock/keystreammock_generated.go -package=keystreammock -source=keymanager.go

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
		mutex         sync.RWMutex
		network       network.Network
		objectStore   *sqlobjectstore.Store
		streamManager stream.Manager
		configStore   configstore.Store
		controller    Controller
		topic         *pubsub.Topic[Controller]
	}
)

func NewKeyManager(
	net network.Network,
	objectStore *sqlobjectstore.Store,
	streamManager stream.Manager,
	configStore configstore.Store,
) (Manager, error) {
	m := &manager{
		network:       net,
		objectStore:   objectStore,
		streamManager: streamManager,
		configStore:   configStore,
		topic:         pubsub.NewTopic[Controller](),
	}

	// find controller from config
	streamRoot, err := configStore.Get(configstore.ConfigKeyManagerController)
	if err == nil && streamRoot != "" {
		// load the controller
		streamController, err := streamManager.GetController(
			tilde.Digest(streamRoot),
		)
		if err != nil {
			return nil, fmt.Errorf("could not load stream: %w", err)
		}
		streamRoot := streamController.GetStreamRoot()
		if streamRoot.IsEmpty() {
			return nil, fmt.Errorf("stream root is empty")
		}
		c, err := RestoreController(
			streamController,
			objectStore,
		)
		if err != nil {
			return nil, err
		}
		m.controller = c
		m.topic.Publish(c)
	}

	return m, nil
}

// NewController creates a new controller in the manager's objectstore
func (m *manager) NewController(
	delegatorSeal *DelegatorSeal,
) (Controller, error) {
	// create controller
	c, err := NewController(
		m.network.GetConnectionInfo().Owner,
		m.objectStore,
		m.streamManager,
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
	m.mutex.RLock()
	c := m.controller
	m.mutex.RUnlock()
	if c != nil {
		return c, nil
	}
	select {
	case c := <-m.topic.Subscribe():
		return c, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
