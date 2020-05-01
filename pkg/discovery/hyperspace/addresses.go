package hyperspace

import (
	"sync"
)

type (
	networkAddresses struct {
		m sync.Map
	}
)

// Put -
func (m *networkAddresses) Put(k string) {
	m.m.Store(k, true)
}

// Delete -
func (m *networkAddresses) Delete(k string) {
	m.m.Delete(k)
}

// List -
func (m *networkAddresses) List() []string {
	hs := []string{}
	m.m.Range(func(k, v interface{}) bool {
		hs = append(hs, k.(string))
		return true
	})
	return hs
}
