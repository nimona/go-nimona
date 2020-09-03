package networkmock

import (
	"sync"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

type (
	MockNetworkSimple struct {
		mutex           sync.Mutex
		SubscribeCalled int
		SubscribeCalls  []network.EnvelopeSubscription
		SendCalled      int
		SendCalls       []error
		ReturnAddresses []string
		ReturnLocalPeer localpeer.LocalPeer
	}
)

func (m *MockNetworkSimple) Subscribe(
	filters ...network.EnvelopeFilter,
) network.EnvelopeSubscription {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.SubscribeCalled >= len(m.SubscribeCalls) {
		panic("too many calls to subscribe")
	}
	r := m.SubscribeCalls[m.SubscribeCalled]
	m.SubscribeCalled++
	return r
}

func (m *MockNetworkSimple) Send(
	ctx context.Context,
	obj object.Object,
	rec *peer.Peer,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.SendCalled >= len(m.SendCalls) {
		panic("too many calls to send")
	}
	r := m.SendCalls[m.SendCalled]
	m.SendCalled++
	return r
}

func (m *MockNetworkSimple) Addresses() []string {
	return m.ReturnAddresses
}

func (m *MockNetworkSimple) Listen(
	ctx context.Context,
	bindAddress string,
) (net.Listener, error) {
	panic("not implemented")
}

func (m *MockNetworkSimple) LocalPeer() localpeer.LocalPeer {
	return m.ReturnLocalPeer
}
