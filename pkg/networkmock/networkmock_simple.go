package networkmock

import (
	"sync"
	"sync/atomic"

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
		SubscribeCalled int32
		SubscribeCalls  []network.EnvelopeSubscription
		SendCalled      int32
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
	subscribeCalled := atomic.LoadInt32(&m.SubscribeCalled)
	if int(subscribeCalled) >= len(m.SubscribeCalls) {
		panic("too many calls to subscribe")
	}
	r := m.SubscribeCalls[m.SubscribeCalled]
	atomic.AddInt32(&m.SubscribeCalled, 1)
	return r
}

func (m *MockNetworkSimple) Send(
	ctx context.Context,
	obj object.Object,
	rec *peer.Peer,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	sendCalled := atomic.LoadInt32(&m.SendCalled)
	if int(sendCalled) >= len(m.SendCalls) {
		panic("too many calls to send")
	}
	r := m.SendCalls[m.SendCalled]
	atomic.AddInt32(&m.SendCalled, 1)
	return r
}

func (m *MockNetworkSimple) Addresses() []string {
	return m.ReturnAddresses
}

func (m *MockNetworkSimple) Listen(
	ctx context.Context,
	bindAddress string,
	options ...network.ListenOption,
) (net.Listener, error) {
	panic("not implemented")
}

func (m *MockNetworkSimple) LocalPeer() localpeer.LocalPeer {
	return m.ReturnLocalPeer
}
