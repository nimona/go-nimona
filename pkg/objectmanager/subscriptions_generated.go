// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/geoah/genny

package objectmanager

import (
	"sync"

	"nimona.io/pkg/stream"
	"nimona.io/pkg/tilde"
)

type (
	subscriptions string // nolint
	// SubscriptionsMap -
	SubscriptionsMap struct {
		m sync.Map
	}
)

// NewSubscriptionsMap constructs a new SyncMap
func NewSubscriptionsMap() *SubscriptionsMap {
	return &SubscriptionsMap{}
}

// GetOrPut -
func (m *SubscriptionsMap) GetOrPut(k tilde.Digest, v *stream.Subscription) (*stream.Subscription, bool) {
	nv, ok := m.m.LoadOrStore(k, v)
	return nv.(*stream.Subscription), ok
}

// Put -
func (m *SubscriptionsMap) Put(k tilde.Digest, v *stream.Subscription) {
	m.m.Store(k, v)
}

// Get -
func (m *SubscriptionsMap) Get(k tilde.Digest) (*stream.Subscription, bool) {
	i, ok := m.m.Load(k)
	if !ok {
		return nil, false
	}

	v, ok := i.(*stream.Subscription)
	if !ok {
		return nil, false
	}

	return v, true
}

// Delete -
func (m *SubscriptionsMap) Delete(k tilde.Digest) {
	m.m.Delete(k)
}

// Range -
func (m *SubscriptionsMap) Range(i func(k tilde.Digest, v *stream.Subscription) bool) {
	m.m.Range(func(k, v interface{}) bool {
		return i(k.(tilde.Digest), v.(*stream.Subscription))
	})
}

// ListKeys -
func (m *SubscriptionsMap) ListKeys() []tilde.Digest {
	vs := []tilde.Digest{}
	m.m.Range(func(k, v interface{}) bool {
		vs = append(vs, k.(tilde.Digest))
		return true
	})
	return vs
}

// ListValues -
func (m *SubscriptionsMap) ListValues() []*stream.Subscription {
	vs := []*stream.Subscription{}
	m.m.Range(func(k, v interface{}) bool {
		vs = append(vs, v.(*stream.Subscription))
		return true
	})
	return vs
}
