// Code generated by MockGen. DO NOT EDIT.
// Source: objectmanager.go

// Package objectmanagermock is a generated GoMock package.
package objectmanagermock

import (
	gomock "github.com/golang/mock/gomock"
	context "nimona.io/pkg/context"
	object "nimona.io/pkg/object"
	objectmanager "nimona.io/pkg/objectmanager"
	peer "nimona.io/pkg/peer"
	reflect "reflect"
	time "time"
)

// MockObjectManager is a mock of ObjectManager interface
type MockObjectManager struct {
	ctrl     *gomock.Controller
	recorder *MockObjectManagerMockRecorder
}

// MockObjectManagerMockRecorder is the mock recorder for MockObjectManager
type MockObjectManagerMockRecorder struct {
	mock *MockObjectManager
}

// NewMockObjectManager creates a new mock instance
func NewMockObjectManager(ctrl *gomock.Controller) *MockObjectManager {
	mock := &MockObjectManager{ctrl: ctrl}
	mock.recorder = &MockObjectManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockObjectManager) EXPECT() *MockObjectManagerMockRecorder {
	return m.recorder
}

// Put mocks base method
func (m *MockObjectManager) Put(ctx context.Context, o object.Object) (object.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", ctx, o)
	ret0, _ := ret[0].(object.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Put indicates an expected call of Put
func (mr *MockObjectManagerMockRecorder) Put(ctx, o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockObjectManager)(nil).Put), ctx, o)
}

// Request mocks base method
func (m *MockObjectManager) Request(ctx context.Context, hash object.Hash, peer *peer.Peer, excludeNested bool) (*object.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Request", ctx, hash, peer, excludeNested)
	ret0, _ := ret[0].(*object.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Request indicates an expected call of Request
func (mr *MockObjectManagerMockRecorder) Request(ctx, hash, peer, excludeNested interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Request", reflect.TypeOf((*MockObjectManager)(nil).Request), ctx, hash, peer, excludeNested)
}

// RequestStream mocks base method
func (m *MockObjectManager) RequestStream(ctx context.Context, rootHash object.Hash, recipients ...*peer.Peer) (object.ReadCloser, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, rootHash}
	for _, a := range recipients {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RequestStream", varargs...)
	ret0, _ := ret[0].(object.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestStream indicates an expected call of RequestStream
func (mr *MockObjectManagerMockRecorder) RequestStream(ctx, rootHash interface{}, recipients ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, rootHash}, recipients...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestStream", reflect.TypeOf((*MockObjectManager)(nil).RequestStream), varargs...)
}

// Subscribe mocks base method
func (m *MockObjectManager) Subscribe(lookupOptions ...objectmanager.LookupOption) objectmanager.ObjectSubscription {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range lookupOptions {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(objectmanager.ObjectSubscription)
	return ret0
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockObjectManagerMockRecorder) Subscribe(lookupOptions ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockObjectManager)(nil).Subscribe), lookupOptions...)
}

// RegisterType mocks base method
func (m *MockObjectManager) RegisterType(objectType string, ttl time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterType", objectType, ttl)
}

// RegisterType indicates an expected call of RegisterType
func (mr *MockObjectManagerMockRecorder) RegisterType(objectType, ttl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterType", reflect.TypeOf((*MockObjectManager)(nil).RegisterType), objectType, ttl)
}
