package keystream

import (
	"sync"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/tilde"
)

const (
	ErrControllerNotFound = errors.Error("keystream: controller not found")
)

type (
	Manager interface {
		ListControllers() []tilde.Digest
		GetController(tilde.Digest) (Controller, error)
		AddController(Controller)
	}
	manager struct {
		mutex       sync.RWMutex
		controllers map[tilde.Digest]Controller
	}
)

func NewKeyManager() Manager {
	return &manager{
		controllers: make(map[tilde.Digest]Controller),
	}
}

// ListControllers returns a list of all the controllers' digests
func (m *manager) ListControllers() []tilde.Digest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	ds := []tilde.Digest{}
	for d := range m.controllers {
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
