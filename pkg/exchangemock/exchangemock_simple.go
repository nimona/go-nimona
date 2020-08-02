package exchangemock

import (
	"sync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

type (
	MockExchangeSimple struct {
		mutex           sync.Mutex
		subscribeCalled int
		SubscribeCalls  []exchange.EnvelopeSubscription
		sendCalled      int
		SendCalls       []error
	}
)

func (m *MockExchangeSimple) Subscribe(
	filters ...exchange.EnvelopeFilter,
) exchange.EnvelopeSubscription {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.subscribeCalled >= len(m.SubscribeCalls) {
		panic("too many calls to subscribe")
	}
	r := m.SubscribeCalls[m.subscribeCalled]
	m.subscribeCalled++
	return r
}

func (m *MockExchangeSimple) Send(
	ctx context.Context,
	obj object.Object,
	rec *peer.Peer,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.sendCalled >= len(m.SendCalls) {
		panic("too many calls to send")
	}
	r := m.SendCalls[m.sendCalled]
	m.sendCalled++
	return r
}
