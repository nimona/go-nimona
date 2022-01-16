package networkmock

import (
	"sync"
	"sync/atomic"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/did"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

var _ network.Network = (*MockNetworkSimple)(nil)

type (
	MockNetworkSimple struct {
		mutex                sync.Mutex
		subscribeCalled      int32
		SubscribeCalls       []network.EnvelopeSubscription
		SubscribeOnceCalled  int32
		SubscribeOnceCalls   []*network.Envelope
		sendCalled           int32
		SendCalls            []error
		ReturnAddresses      []string
		ReturnPeerKey        crypto.PrivateKey
		ReturnConnectionInfo *peer.ConnectionInfo
		ReturnRelays         []*peer.ConnectionInfo
	}
)

func (m *MockNetworkSimple) Subscribe(
	filters ...network.EnvelopeFilter,
) network.EnvelopeSubscription {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	subscribeCalled := atomic.LoadInt32(&m.subscribeCalled)
	if int(subscribeCalled) >= len(m.SubscribeCalls) {
		panic("too many calls to subscribe")
	}
	r := m.SubscribeCalls[m.subscribeCalled]
	atomic.AddInt32(&m.subscribeCalled, 1)
	return r
}

func (m *MockNetworkSimple) SubscribeOnce(
	ctx context.Context,
	filters ...network.EnvelopeFilter,
) (*network.Envelope, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	r := m.SubscribeOnceCalls[m.subscribeCalled]
	m.subscribeCalled++
	return r, nil
}

func (m *MockNetworkSimple) Send(
	ctx context.Context,
	obj *object.Object,
	id did.DID,
	opt ...network.SendOption,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	sendCalled := atomic.LoadInt32(&m.sendCalled)
	if int(sendCalled) >= len(m.SendCalls) {
		panic("too many calls to send")
	}
	r := m.SendCalls[m.sendCalled]
	atomic.AddInt32(&m.sendCalled, 1)
	return r
}

func (m *MockNetworkSimple) SendCalled() int {
	sendCalled := atomic.LoadInt32(&m.sendCalled)
	return int(sendCalled)
}

func (m *MockNetworkSimple) SubscribeCalled() int {
	subscribeCalled := atomic.LoadInt32(&m.subscribeCalled)
	return int(subscribeCalled)
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

func (m *MockNetworkSimple) RegisterResolver(
	resolver network.Resolver,
) {
}

func (m *MockNetworkSimple) GetPeerKey() crypto.PrivateKey {
	return m.ReturnPeerKey
}

func (m *MockNetworkSimple) GetAddresses() []string {
	return m.ReturnAddresses
}

func (m *MockNetworkSimple) RegisterAddresses(addresses ...string) {
}

func (m *MockNetworkSimple) GetConnectionInfo() *peer.ConnectionInfo {
	return m.ReturnConnectionInfo
}

func (m *MockNetworkSimple) GetRelays() []*peer.ConnectionInfo {
	return m.ReturnRelays
}

func (m *MockNetworkSimple) RegisterRelays(relays ...*peer.ConnectionInfo) {
}

func (m *MockNetworkSimple) Close() error {
	return nil
}
