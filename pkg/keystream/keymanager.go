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
	KeyStreamManager interface {
		ListControllers() []tilde.Digest
		GetController(tilde.Digest) (Controller, error)
		AddController(Controller)
	}
	keyStreamManager struct {
		mutex       sync.RWMutex
		controllers map[tilde.Digest]Controller
	}
)

func NewKeyManager() KeyStreamManager {
	return &keyStreamManager{
		controllers: make(map[tilde.Digest]Controller),
	}
}

// ListControllers returns a list of all the controllers' digests
func (m *keyStreamManager) ListControllers() []tilde.Digest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	ds := []tilde.Digest{}
	for d := range m.controllers {
		ds = append(ds, d)
	}
	return ds
}

func (m *keyStreamManager) GetController(d tilde.Digest) (Controller, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	c, ok := m.controllers[d]
	if !ok {
		return nil, ErrControllerNotFound
	}
	return c, nil
}

func (m *keyStreamManager) AddController(c Controller) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.controllers[c.GetKeyStream().Root] = c
}
