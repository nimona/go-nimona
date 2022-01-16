// Code generated by MockGen. DO NOT EDIT.
// Source: network.go

// Package networkmock is a generated GoMock package.
package networkmock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	net "nimona.io/internal/net"
	context "nimona.io/pkg/context"
	crypto "nimona.io/pkg/crypto"
	did "nimona.io/pkg/did"
	network "nimona.io/pkg/network"
	object "nimona.io/pkg/object"
	peer "nimona.io/pkg/peer"
)

// MockResolver is a mock of Resolver interface.
type MockResolver struct {
	ctrl     *gomock.Controller
	recorder *MockResolverMockRecorder
}

// MockResolverMockRecorder is the mock recorder for MockResolver.
type MockResolverMockRecorder struct {
	mock *MockResolver
}

// NewMockResolver creates a new mock instance.
func NewMockResolver(ctrl *gomock.Controller) *MockResolver {
	mock := &MockResolver{ctrl: ctrl}
	mock.recorder = &MockResolverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResolver) EXPECT() *MockResolverMockRecorder {
	return m.recorder
}

// LookupPeer mocks base method.
func (m *MockResolver) LookupPeer(ctx context.Context, id did.DID) ([]*peer.ConnectionInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LookupPeer", ctx, id)
	ret0, _ := ret[0].([]*peer.ConnectionInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LookupPeer indicates an expected call of LookupPeer.
func (mr *MockResolverMockRecorder) LookupPeer(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LookupPeer", reflect.TypeOf((*MockResolver)(nil).LookupPeer), ctx, id)
}

// MockNetwork is a mock of Network interface.
type MockNetwork struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkMockRecorder
}

// MockNetworkMockRecorder is the mock recorder for MockNetwork.
type MockNetworkMockRecorder struct {
	mock *MockNetwork
}

// NewMockNetwork creates a new mock instance.
func NewMockNetwork(ctrl *gomock.Controller) *MockNetwork {
	mock := &MockNetwork{ctrl: ctrl}
	mock.recorder = &MockNetworkMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNetwork) EXPECT() *MockNetworkMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockNetwork) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockNetworkMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockNetwork)(nil).Close))
}

// GetAddresses mocks base method.
func (m *MockNetwork) GetAddresses() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddresses")
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetAddresses indicates an expected call of GetAddresses.
func (mr *MockNetworkMockRecorder) GetAddresses() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddresses", reflect.TypeOf((*MockNetwork)(nil).GetAddresses))
}

// GetConnectionInfo mocks base method.
func (m *MockNetwork) GetConnectionInfo() *peer.ConnectionInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConnectionInfo")
	ret0, _ := ret[0].(*peer.ConnectionInfo)
	return ret0
}

// GetConnectionInfo indicates an expected call of GetConnectionInfo.
func (mr *MockNetworkMockRecorder) GetConnectionInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConnectionInfo", reflect.TypeOf((*MockNetwork)(nil).GetConnectionInfo))
}

// GetPeerKey mocks base method.
func (m *MockNetwork) GetPeerKey() crypto.PrivateKey {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPeerKey")
	ret0, _ := ret[0].(crypto.PrivateKey)
	return ret0
}

// GetPeerKey indicates an expected call of GetPeerKey.
func (mr *MockNetworkMockRecorder) GetPeerKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPeerKey", reflect.TypeOf((*MockNetwork)(nil).GetPeerKey))
}

// GetRelays mocks base method.
func (m *MockNetwork) GetRelays() []*peer.ConnectionInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRelays")
	ret0, _ := ret[0].([]*peer.ConnectionInfo)
	return ret0
}

// GetRelays indicates an expected call of GetRelays.
func (mr *MockNetworkMockRecorder) GetRelays() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRelays", reflect.TypeOf((*MockNetwork)(nil).GetRelays))
}

// Listen mocks base method.
func (m *MockNetwork) Listen(ctx context.Context, bindAddress string, options ...network.ListenOption) (net.Listener, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, bindAddress}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Listen", varargs...)
	ret0, _ := ret[0].(net.Listener)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Listen indicates an expected call of Listen.
func (mr *MockNetworkMockRecorder) Listen(ctx, bindAddress interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, bindAddress}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Listen", reflect.TypeOf((*MockNetwork)(nil).Listen), varargs...)
}

// RegisterAddresses mocks base method.
func (m *MockNetwork) RegisterAddresses(arg0 ...string) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "RegisterAddresses", varargs...)
}

// RegisterAddresses indicates an expected call of RegisterAddresses.
func (mr *MockNetworkMockRecorder) RegisterAddresses(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterAddresses", reflect.TypeOf((*MockNetwork)(nil).RegisterAddresses), arg0...)
}

// RegisterRelays mocks base method.
func (m *MockNetwork) RegisterRelays(arg0 ...*peer.ConnectionInfo) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "RegisterRelays", varargs...)
}

// RegisterRelays indicates an expected call of RegisterRelays.
func (mr *MockNetworkMockRecorder) RegisterRelays(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterRelays", reflect.TypeOf((*MockNetwork)(nil).RegisterRelays), arg0...)
}

// RegisterResolver mocks base method.
func (m *MockNetwork) RegisterResolver(resolver network.Resolver) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterResolver", resolver)
}

// RegisterResolver indicates an expected call of RegisterResolver.
func (mr *MockNetworkMockRecorder) RegisterResolver(resolver interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterResolver", reflect.TypeOf((*MockNetwork)(nil).RegisterResolver), resolver)
}

// Send mocks base method.
func (m *MockNetwork) Send(ctx context.Context, object *object.Object, id did.DID, sendOptions ...network.SendOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, object, id}
	for _, a := range sendOptions {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Send", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockNetworkMockRecorder) Send(ctx, object, id interface{}, sendOptions ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, object, id}, sendOptions...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockNetwork)(nil).Send), varargs...)
}

// Subscribe mocks base method.
func (m *MockNetwork) Subscribe(filters ...network.EnvelopeFilter) network.EnvelopeSubscription {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range filters {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(network.EnvelopeSubscription)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockNetworkMockRecorder) Subscribe(filters ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockNetwork)(nil).Subscribe), filters...)
}

// SubscribeOnce mocks base method.
func (m *MockNetwork) SubscribeOnce(ctx context.Context, filters ...network.EnvelopeFilter) (*network.Envelope, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range filters {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SubscribeOnce", varargs...)
	ret0, _ := ret[0].(*network.Envelope)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscribeOnce indicates an expected call of SubscribeOnce.
func (mr *MockNetworkMockRecorder) SubscribeOnce(ctx interface{}, filters ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, filters...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeOnce", reflect.TypeOf((*MockNetwork)(nil).SubscribeOnce), varargs...)
}
