// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/geoah/genny

package peer

import (
	"sync"
)

type (
	// StringAddressesSyncMap -
	StringAddressesSyncMap struct {
		m sync.Map
	}
)

// NewStringAddressesSyncMap constructs a new SyncMap
func NewStringAddressesSyncMap() *StringAddressesSyncMap {
	return &StringAddressesSyncMap{}
}

// Put -
func (m *StringAddressesSyncMap) Put(k string, v *Addresses) {
	m.m.Store(k, v)
}

// Get -
func (m *StringAddressesSyncMap) Get(k string) (*Addresses, bool) {
	i, ok := m.m.Load(k)
	if !ok {
		return nil, false
	}

	v, ok := i.(*Addresses)
	if !ok {
		return nil, false
	}

	return v, true
}

// Delete -
func (m *StringAddressesSyncMap) Delete(k string) {
	m.m.Delete(k)
}

// Range -
func (m *StringAddressesSyncMap) Range(i func(k string, v *Addresses) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return i(k.(string), v.(*Addresses))
	})
}
