// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package exchange

import "sync"

type (
	// StringSendRequestSyncMap -
	StringSendRequestSyncMap struct {
		m sync.Map
	}
)

// NewStringSendRequestSyncMap constructs a new SyncMap
func NewStringSendRequestSyncMap() *StringSendRequestSyncMap {
	return &StringSendRequestSyncMap{}
}

// Put -
func (m *StringSendRequestSyncMap) Put(k string, v *sendRequest) {
	m.m.Store(k, v)
}

// Get -
func (m *StringSendRequestSyncMap) Get(k string) (*sendRequest, bool) {
	i, ok := m.m.Load(k)
	if !ok {
		return nil, false
	}

	v, ok := i.(*sendRequest)
	if !ok {
		return nil, false
	}

	return v, true
}

// Delete -
func (m *StringSendRequestSyncMap) Delete(k string) {
	m.m.Delete(k)
}
