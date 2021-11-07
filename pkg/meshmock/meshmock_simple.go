package meshmock

import (
	"sync"
	"sync/atomic"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/mesh"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

var _ mesh.Mesh = (*MockMeshSimple)(nil)

type (
	MockMeshSimple struct {
		mutex                sync.Mutex
		subscribeCalled      int32
		SubscribeCalls       []mesh.EnvelopeSubscription
		SubscribeOnceCalled  int32
		SubscribeOnceCalls   []*mesh.Envelope
		sendCalled           int32
		SendCalls            []error
		ReturnAddresses      []string
		ReturnPeerKey        crypto.PrivateKey
		ReturnConnectionInfo *peer.ConnectionInfo
		ReturnRelays         []*peer.ConnectionInfo
	}
)

func (m *MockMeshSimple) Subscribe(
	filters ...mesh.EnvelopeFilter,
) mesh.EnvelopeSubscription {
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

func (m *MockMeshSimple) SubscribeOnce(
	ctx context.Context,
	filters ...mesh.EnvelopeFilter,
) (*mesh.Envelope, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	r := m.SubscribeOnceCalls[m.subscribeCalled]
	m.subscribeCalled++
	return r, nil
}

func (m *MockMeshSimple) Send(
	ctx context.Context,
	obj *object.Object,
	rec crypto.PublicKey,
	opt ...mesh.SendOption,
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

func (m *MockMeshSimple) SendCalled() int {
	sendCalled := atomic.LoadInt32(&m.sendCalled)
	return int(sendCalled)
}

func (m *MockMeshSimple) SubscribeCalled() int {
	subscribeCalled := atomic.LoadInt32(&m.subscribeCalled)
	return int(subscribeCalled)
}

func (m *MockMeshSimple) Addresses() []string {
	return m.ReturnAddresses
}

func (m *MockMeshSimple) Listen(
	ctx context.Context,
	bindAddress string,
	options ...mesh.ListenOption,
) (net.Listener, error) {
	panic("not implemented")
}

func (m *MockMeshSimple) RegisterResolver(
	resolver mesh.Resolver,
) {
}

func (m *MockMeshSimple) GetPeerKey() crypto.PrivateKey {
	return m.ReturnPeerKey
}

func (m *MockMeshSimple) GetAddresses() []string {
	return m.ReturnAddresses
}

func (m *MockMeshSimple) RegisterAddresses(addresses ...string) {
}

func (m *MockMeshSimple) GetConnectionInfo() *peer.ConnectionInfo {
	return m.ReturnConnectionInfo
}

func (m *MockMeshSimple) GetRelays() []*peer.ConnectionInfo {
	return m.ReturnRelays
}

func (m *MockMeshSimple) RegisterRelays(relays ...*peer.ConnectionInfo) {
}

func (m *MockMeshSimple) Close() error {
	return nil
}
